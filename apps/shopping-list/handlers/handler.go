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
		var checked int
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &checked); err != nil {
			continue
		}
		item.Checked = checked == 1
		items = append(items, item)
	}

	return items, nil
}

func (h *Handler) homeContent(stores []Store) string {
	content := `
		<div class="px-4 py-8">
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-2xl font-bold text-gray-900">My Shopping Lists</h2>
				<button 
					hx-get="/stores/new" 
					hx-target="#modal"
					class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700"
				>
					+ New Store
				</button>
			</div>

			<div id="stores-container" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
	`

	if len(stores) == 0 {
		content += `
				<div class="col-span-full text-center py-12 text-gray-500">
					<p class="text-lg">No stores yet. Create your first shopping list!</p>
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
				<div class="bg-white rounded-lg shadow p-6 border-t-4" style="border-color: %s">
					<div class="flex justify-between items-start mb-4">
						<h3 class="text-xl font-bold text-gray-900">%s</h3>
						<button 
							hx-delete="/stores/%d" 
							hx-confirm="Delete this store and all items?"
							hx-target="#stores-container"
							hx-swap="outerHTML"
							class="text-red-600 hover:text-red-800"
						>
							🗑️
						</button>
					</div>
					
					<div class="mb-4 text-sm text-gray-600">
						%d items to buy
					</div>

					<form hx-post="/stores/%d/items" hx-target="#store-%d-items" hx-swap="beforeend" class="mb-4">
						<div class="flex gap-2">
							<input 
								type="text" 
								name="name" 
								placeholder="Add item..." 
								required
								class="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
							/>
							<input 
								type="text" 
								name="quantity" 
								placeholder="Qty" 
								class="w-20 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
							/>
							<button type="submit" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700">
								+
							</button>
						</div>
					</form>

					<div id="store-%d-items" class="space-y-2">
		`, store.Color, store.Name, store.ID, uncheckedCount, store.ID, store.ID, store.ID)

		for _, item := range store.Items {
			checkedClass := ""
			if item.Checked {
				checkedClass = "line-through text-gray-400"
			}

			content += fmt.Sprintf(`
						<div class="flex items-center gap-2 p-2 hover:bg-gray-50 rounded">
							<input 
								type="checkbox" 
								%s
								hx-post="/items/%d/toggle"
								hx-target="closest div"
								hx-swap="outerHTML"
								class="h-5 w-5 text-indigo-600 rounded"
							/>
							<span class="flex-1 %s">%s</span>
							<span class="text-sm text-gray-500 %s">%s</span>
							<button 
								hx-delete="/items/%d"
								hx-target="closest div"
								hx-swap="outerHTML"
								class="text-red-600 hover:text-red-800 text-sm"
							>
								✕
							</button>
						</div>
			`, map[bool]string{true: "checked", false: ""}[item.Checked], item.ID, checkedClass, item.Name, checkedClass, item.Quantity, item.ID)
		}

		content += `
					</div>
				</div>
		`
	}

	content += `
			</div>
		</div>

		<div id="modal"></div>
	`

	return content
}
