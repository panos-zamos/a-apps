package handlers

import (
	"html/template"
	"net/http"

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
}

// LoginPage renders the login page
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("login").Parse(sharedTemplates.LoginHTML))
	data := map[string]interface{}{
		"AppName": "todo-list",
		"Error":   r.URL.Query().Get("error"),
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
		"Title":    "Todo List",
		"AppName":  "Todo List",
		"Username": username,
		"Content":  template.HTML(h.homeContent(stores)),
	}

	tmpl := template.Must(template.New("base").Parse(sharedTemplates.BaseHTML))
	tmpl.Execute(w, data)
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
}

func (h *Handler) getStores(username string) ([]Store, error) {
	rows, err := h.DB.Query(`
		SELECT id, name, color FROM stores 
		WHERE username = ? 
		ORDER BY name
	`, username)
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
		s.Items, _ = h.getItems(s.ID, username)
		stores = append(stores, s)
	}

	return stores, nil
}

func (h *Handler) getItems(storeID int, username string) ([]Item, error) {
	rows, err := h.DB.Query(`
		SELECT id, name, quantity, checked 
		FROM items 
		WHERE store_id = ? AND username = ? 
		ORDER BY checked ASC, created_at DESC
	`, storeID, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var checked bool
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &checked); err != nil {
			continue
		}
		item.Checked = checked
		items = append(items, item)
	}

	return items, nil
}

func (h *Handler) homeContent(stores []Store) string {
	content := `
		<div class="row space-between mb-md">
			<h2>Shopping Lists</h2>
			<button class="btn-add" hx-get="/stores/new" hx-target="#modal">+</button>
		</div>
	`

	content += h.storesGrid(stores)
	content += `<div id="modal" class="mt-lg"></div>`
	return content
}
