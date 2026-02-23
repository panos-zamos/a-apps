package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/panos/a-apps/shared/auth"
	"github.com/panos/a-apps/shared/database"
	"github.com/panos/a-apps/shared/models"
	sharedTemplates "github.com/panos/a-apps/shared/templates"
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
		"AppName": "shopping-list",
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
		"Title":    "Shopping List",
		"AppName":  "Shopping List",
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
		<section class="mb-lg">
			<div class="row space-between mb-md">
				<h2>My Shopping Lists</h2>
				<button 
					class="primary"
					hx-get="/stores/new" 
					hx-target="#modal"
				>
					New Store
				</button>
			</div>

			<div id="stores-container">
	`

	if len(stores) == 0 {
		content += `
				<div class="panel center">
					<p class="muted">No stores yet. Create your first shopping list!</p>
				</div>
		`
	}

	for _, store := range stores {
		uncheckedCount := 0
		for _, item := range store.Items {
			if !item.Checked {
				uncheckedCount++
			}
		}

		content += fmt.Sprintf(`
				<article class="panel mb-md">
					<div class="row space-between mb-sm">
						<h3>%s</h3>
						<button 
							hx-delete="/stores/%d" 
							hx-confirm="Delete this store and all items?"
							hx-target="#stores-container"
							hx-swap="outerHTML"
						>
							Delete
						</button>
					</div>
					
					<p class="muted">%d items to buy</p>

					<form hx-post="/stores/%d/items" hx-target="#store-%d-items" hx-swap="beforeend" class="mt-md">
						<label>Item</label>
						<input 
							type="text" 
							name="name" 
							placeholder="Add item..." 
							required
						/>
						<div class="mt-sm">
							<label>Quantity</label>
							<input 
								type="text" 
								name="quantity" 
								placeholder="Qty" 
							/>
						</div>
						<div class="mt-md">
							<button type="submit" class="primary">Add Item</button>
						</div>
					</form>

					<div id="store-%d-items" class="mt-md">
		`, store.Name, store.ID, uncheckedCount, store.ID, store.ID, store.ID)

		for _, item := range store.Items {
			checkedAttr := ""
			itemName := item.Name
			itemQuantity := item.Quantity
			if item.Checked {
				checkedAttr = "checked"
				itemName = fmt.Sprintf("<del>%s</del>", item.Name)
				itemQuantity = fmt.Sprintf("<del>%s</del>", item.Quantity)
			}

			content += fmt.Sprintf(`
						<div class="row space-between mb-sm">
							<div class="row">
								<input 
									type="checkbox" 
									%s
									hx-post="/items/%d/toggle"
									hx-target="closest .space-between"
									hx-swap="outerHTML"
								/>
								<span>%s</span>
								<span class="muted">%s</span>
							</div>
							<button 
								hx-delete="/items/%d"
								hx-target="closest .space-between"
								hx-swap="outerHTML"
							>
								Remove
							</button>
						</div>
			`, checkedAttr, item.ID, itemName, itemQuantity, item.ID)
		}

		content += `
					</div>
				</article>
		`
	}

	content += `
			</div>
		</section>

		<div id="modal" class="mt-lg"></div>
	`

	return content
}
