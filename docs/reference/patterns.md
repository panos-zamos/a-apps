# Common Patterns and Code Templates

This document contains reusable patterns for building apps in this monorepo.

## Database Patterns

### Creating a Table with User Isolation

```go
`CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    username TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`
```

### Foreign Key Relationship

```go
`CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES parents(id) ON DELETE CASCADE
)`
```

### Indexes for Performance

```go
`CREATE INDEX IF NOT EXISTS idx_items_username ON items(username)`,
`CREATE INDEX IF NOT EXISTS idx_items_parent ON items(parent_id)`,
```

### Query with User Filter

```go
rows, err := h.DB.Query(`
    SELECT id, name, created_at 
    FROM items 
    WHERE username = ? 
    ORDER BY created_at DESC
`, username)
```

## HTMX Patterns

### Form that Adds Item to List

```go
// In homeContent() or similar:
html := fmt.Sprintf(`
    <form hx-post="/items" hx-target="#items-list" hx-swap="beforeend">
        <input type="text" name="name" placeholder="Item name" required />
        <button type="submit">Add</button>
    </form>
    
    <div id="items-list">
        <!-- Items will be added here -->
    </div>
`)
```

### Handler Returns HTML Fragment

```go
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
    username, _ := auth.GetUsername(r)
    name := r.FormValue("name")
    
    result, err := h.DB.Exec(`
        INSERT INTO items (name, username) VALUES (?, ?)
    `, name, username)
    
    if err != nil {
        http.Error(w, "Failed to create", http.StatusInternalServerError)
        return
    }
    
    itemID, _ := result.LastInsertId()
    
    // Return HTML for the new item
    html := fmt.Sprintf(`
        <div class="item">
            <span>%s</span>
            <button hx-delete="/items/%d" hx-target="closest div">Delete</button>
        </div>
    `, name, itemID)
    
    w.Write([]byte(html))
}
```

### Toggle/Checkbox Action

```go
// In HTML:
`<input type="checkbox" hx-post="/items/%d/toggle" hx-target="closest div" hx-swap="outerHTML" />`

// Handler:
func (h *Handler) ToggleItem(w http.ResponseWriter, r *http.Request) {
    itemID := chi.URLParam(r, "id")
    
    _, err := h.DB.Exec(`UPDATE items SET checked = NOT checked WHERE id = ?`, itemID)
    if err != nil {
        http.Error(w, "Failed", http.StatusInternalServerError)
        return
    }
    
    // Get updated item and return full HTML
    var checked int
    var name string
    h.DB.QueryRow(`SELECT name, checked FROM items WHERE id = ?`, itemID).Scan(&name, &checked)
    
    checkedAttr := ""
    if checked == 1 {
        checkedAttr = "checked"
    }
    
    html := fmt.Sprintf(`
        <div class="item">
            <input type="checkbox" %s hx-post="/items/%s/toggle" hx-target="closest div" hx-swap="outerHTML" />
            <span>%s</span>
        </div>
    `, checkedAttr, itemID, name)
    
    w.Write([]byte(html))
}
```

### Delete Action

```go
// In HTML:
`<button hx-delete="/items/%d" hx-confirm="Delete this item?" hx-target="closest div" hx-swap="outerHTML">Delete</button>`

// Handler:
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
    username, _ := auth.GetUsername(r)
    itemID := chi.URLParam(r, "id")
    
    _, err := h.DB.Exec(`DELETE FROM items WHERE id = ? AND username = ?`, itemID, username)
    if err != nil {
        http.Error(w, "Failed", http.StatusInternalServerError)
        return
    }
    
    // Return empty (element will be removed from DOM)
    w.WriteHeader(http.StatusOK)
}
```

### Modal/Popup Form

```go
// Button to open modal:
`<button hx-get="/items/new" hx-target="#modal">New Item</button>`

// Modal container in HTML:
`<div id="modal"></div>`

// Handler returns modal HTML:
func (h *Handler) NewItemForm(w http.ResponseWriter, r *http.Request) {
    html := `
        <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <div class="bg-white rounded-lg p-6 max-w-md w-full">
                <h3 class="text-xl font-bold mb-4">New Item</h3>
                <form hx-post="/items" hx-target="#modal">
                    <input type="text" name="name" placeholder="Name" required />
                    <button type="submit">Create</button>
                    <button type="button" onclick="this.closest('.fixed').remove()">Cancel</button>
                </form>
            </div>
        </div>
    `
    w.Write([]byte(html))
}
```

## Handler Patterns

### Standard Handler Structure

```go
type Handler struct {
    DB        *database.DB
    Users     []models.UserFromConfig
    JWTSecret string
}
```

### Get Username from Context

```go
func (h *Handler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := auth.GetUsername(r)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Use username for queries...
}
```

### Render Page with Base Template

```go
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
    username, _ := auth.GetUsername(r)
    
    data := map[string]interface{}{
        "Title":    "My App",
        "AppName":  "My App",
        "Username": username,
        "Content":  template.HTML(h.homeContent()),
    }
    
    tmpl := template.Must(template.New("base").Parse(sharedTemplates.BaseHTML))
    tmpl.Execute(w, data)
}

func (h *Handler) homeContent() string {
    // Return HTML string
    return `<div>Content here</div>`
}
```

## Routing Patterns

### Chi Router Setup

```go
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// Public routes
r.Get("/login", h.LoginPage)
r.Post("/login", h.Login)
r.Post("/logout", h.Logout)
r.Get("/health", h.HealthCheck)

// Protected routes
r.Group(func(r chi.Router) {
    r.Use(auth.Middleware(jwtSecret))
    
    r.Get("/", h.Home)
    r.Get("/items/new", h.NewItemForm)
    r.Post("/items", h.CreateItem)
    r.Post("/items/{id}/toggle", h.ToggleItem)
    r.Delete("/items/{id}", h.DeleteItem)
})
```

### URL Parameters

```go
func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
    itemID := chi.URLParam(r, "id")  // From route like /items/{id}
    
    // Use itemID...
}
```

## UI / CSS patterns (UI contract)

This repo uses a strict UI contract.

- **Rules**: see [design-spec.md](./design-spec.md)
- **Do not** introduce Tailwind (or other utility frameworks)
- **Do not** invent new CSS classes; use the classes provided by `custom.css`

### Buttons

```html
<button class="primary">Save</button>
<button>Cancel</button>
<button class="danger">Delete</button>
```

### Panels (grouped content)

```html
<section class="panel">
  <h2>Settings</h2>
  <p class="muted">Secondary text</p>
</section>
```

### Cards (list items)

```html
<div class="card-list">
  <a class="card" href="/items/1">
    <div class="card-header">
      <span class="card-name">Item</span>
      <span class="badge badge-dev">development</span>
    </div>
    <p class="card-desc">Short description</p>
  </a>
</div>
```

### Common layout helpers

```html
<div class="row space-between">
  <span class="muted">Left</span>
  <span>Right</span>
</div>
```

## Migration Patterns

### Add Column to Existing Table

```go
// In new migration:
`ALTER TABLE items ADD COLUMN description TEXT DEFAULT ''`,
```

### Create Junction Table (Many-to-Many)

```go
`CREATE TABLE IF NOT EXISTS item_tags (
    item_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (item_id, tag_id),
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
)`
```

## Testing Patterns

### Basic Handler Test

```go
func TestCreateItem(t *testing.T) {
    // Setup
    db, _ := database.Open(":memory:")
    defer db.Close()
    db.RunMigrations(Migrations)
    
    h := &Handler{
        DB: db,
        Users: []models.UserFromConfig{
            {Username: "test", PasswordHash: "hash"},
        },
        JWTSecret: "test-secret",
    }
    
    // Create request
    req := httptest.NewRequest("POST", "/items", strings.NewReader("name=Test"))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, "test"))
    
    // Execute
    w := httptest.NewRecorder()
    h.CreateItem(w, req)
    
    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

## Docker Patterns

### Multi-stage Build

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
RUN mkdir -p data
EXPOSE 3001
CMD ["./main"]
```

## Configuration Patterns

### Load from YAML

```go
import "gopkg.in/yaml.v3"

type Config struct {
    AppName string `yaml:"app_name"`
    Port    int    `yaml:"port"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### Environment Variables with Fallback

```go
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    jwtSecret = "dev-secret-change-in-production"
}

port := os.Getenv("PORT")
if port == "" {
    port = "3001"
}
```

---

## Quick Reference: HTMX Attributes

- `hx-get="/url"` - GET request
- `hx-post="/url"` - POST request
- `hx-delete="/url"` - DELETE request
- `hx-target="#id"` - Where to put response
- `hx-swap="innerHTML"` - Replace inner HTML (default)
- `hx-swap="outerHTML"` - Replace entire element
- `hx-swap="beforeend"` - Append to end
- `hx-swap="afterbegin"` - Prepend to start
- `hx-confirm="Message"` - Confirmation dialog
- `hx-trigger="click"` - Event that triggers (default)

## Quick Reference: UI contract

When in doubt, follow the UI contract:

- [design-spec.md](./design-spec.md)
