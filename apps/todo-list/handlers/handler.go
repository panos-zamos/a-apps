package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/panos-zamos/a-apps/shared/auth"
	"github.com/panos-zamos/a-apps/shared/database"
	"github.com/panos-zamos/a-apps/shared/models"
	sharedTemplates "github.com/panos-zamos/a-apps/shared/templates"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	DB        *database.DB
	Users     []models.UserFromConfig
	JWTSecret string
	AppConfig models.AppConfig
}

// LoginPage renders the login page
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("login").Parse(sharedTemplates.LoginHTML))
	data := map[string]interface{}{
		"AppName":        "todo-list",
		"AppVersion":     h.AppConfig.AppVersion,
		"AppReleaseDate": h.AppConfig.AppReleaseDate,
		"ChangelogURL":   "/changelog",
		"Error":          r.URL.Query().Get("error"),
	}
	tmpl.Execute(w, data)
}

// Login handles login form submission
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	valid, err := auth.ValidateCredentials(username, password, h.Users)
	if !valid || err != nil {
		http.Redirect(w, r, "/login?error=Invalid credentials", http.StatusSeeOther)
		return
	}

	// Generate token
	token, err := auth.GenerateToken(username, h.JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HealthCheck returns OK for health checks
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Home renders the home page with shopping lists
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)

	// Get all stores for this user
	stores, _ := h.getStores(username)

	data := map[string]interface{}{
		"Title":          "Todo List",
		"AppName":        "Todo List",
		"Username":       username,
		"Content":        template.HTML(h.homeContent(stores)),
		"AppVersion":     h.AppConfig.AppVersion,
		"AppReleaseDate": h.AppConfig.AppReleaseDate,
		"ChangelogURL":   "/changelog",
	}

	tmpl := template.Must(template.New("base").Parse(sharedTemplates.BaseHTML))
	tmpl.Execute(w, data)
}

// ChangelogPage renders the changelog page.
func (h *Handler) ChangelogPage(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	entries, err := models.LoadChangelog(h.changelogPath())

	content := ""
	if err != nil {
		content = `<div class="panel"><p class="muted">Changelog unavailable.</p></div>`
	} else {
		content = h.changelogContent(entries)
	}

	data := map[string]interface{}{
		"Title":          "Changelog",
		"AppName":        "Todo List",
		"Username":       username,
		"Content":        template.HTML(content),
		"AppVersion":     h.AppConfig.AppVersion,
		"AppReleaseDate": h.AppConfig.AppReleaseDate,
		"ChangelogURL":   "/changelog",
	}

	tmpl := template.Must(template.New("base").Parse(sharedTemplates.BaseHTML))
	tmpl.Execute(w, data)
}

func (h *Handler) changelogPath() string {
	if h.AppConfig.ChangelogPath != "" {
		return h.AppConfig.ChangelogPath
	}
	return "changelog.yaml"
}

func (h *Handler) changelogContent(entries []models.ChangelogEntry) string {
	if len(entries) == 0 {
		return `<div class="panel"><p class="muted">No changelog entries yet.</p></div>`
	}

	var builder strings.Builder
	builder.WriteString(`<h2 class="mb-md">changelog</h2>`)

	for _, entry := range entries {
		version := template.HTMLEscapeString(entry.Version)
		date := template.HTMLEscapeString(entry.Date)

		builder.WriteString(`<section class="panel mb-md">`)
		if version != "" {
			builder.WriteString(fmt.Sprintf("<h3>v%s</h3>", version))
		} else {
			builder.WriteString("<h3>Unversioned</h3>")
		}
		if date != "" {
			builder.WriteString(fmt.Sprintf("<p class=\"muted\">%s</p>", date))
		}
		if len(entry.Changes) > 0 {
			builder.WriteString(`<ul class="mt-sm">`)
			for _, change := range entry.Changes {
				builder.WriteString(fmt.Sprintf("<li>%s</li>", template.HTMLEscapeString(change)))
			}
			builder.WriteString("</ul>")
		}
		builder.WriteString("</section>")
	}

	return builder.String()
}

type Store struct {
	ID    int
	Name  string
	Color string
	Items []Item
}

type Item struct {
	ID       int
	Name     string
	Quantity string
	Checked  bool
	Username string
}

func (h *Handler) getStores(username string) ([]Store, error) {
	sharedUsers := h.sharedUsernames(username)
	query := fmt.Sprintf(`
		SELECT id, name, color FROM stores 
		WHERE username IN (%s) 
		ORDER BY name
	`, buildPlaceholders(len(sharedUsers)))
	rows, err := h.DB.Query(query, toInterfaceSlice(sharedUsers)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var s Store
		if err := rows.Scan(&s.ID, &s.Name, &s.Color); err != nil {
			continue
		}

		// Get items for this store
		s.Items, _ = h.getItems(s.ID)
		stores = append(stores, s)
	}

	return stores, nil
}

func (h *Handler) getItems(storeID int) ([]Item, error) {
	rows, err := h.DB.Query(`
		SELECT id, name, quantity, checked, username 
		FROM items 
		WHERE store_id = ? 
		ORDER BY checked ASC, created_at DESC
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var checked bool
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &checked, &item.Username); err != nil {
			continue
		}
		item.Checked = checked
		items = append(items, item)
	}

	return items, nil
}

func (h *Handler) sharedUsernames(username string) []string {
	if username == "" {
		return []string{""}
	}

	group := ""
	for _, user := range h.Users {
		if user.Username == username {
			group = user.ShareGroup
			break
		}
	}

	if group == "" {
		return []string{username}
	}

	shared := make([]string, 0, len(h.Users))
	for _, user := range h.Users {
		if user.ShareGroup == group {
			shared = append(shared, user.Username)
		}
	}

	if len(shared) == 0 {
		return []string{username}
	}

	return shared
}

func buildPlaceholders(count int) string {
	if count <= 0 {
		return "?"
	}

	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}

func toInterfaceSlice(values []string) []interface{} {
	args := make([]interface{}, len(values))
	for i, value := range values {
		args[i] = value
	}
	return args
}

func (h *Handler) homeContent(stores []Store) string {
	content := `
		<div class="row space-between mb-md">
			<h2>lists</h2>
			<button class="btn" hx-get="/stores/new" hx-target="#modal">+ add list</button>
		</div>
	`

	content += h.storesGrid(stores)
	content += `<div id="modal" class="mt-lg"></div>`
	return content
}
