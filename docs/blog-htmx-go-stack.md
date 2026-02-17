# HTMX + Go: The Underrated Stack for Building Personal Apps Fast

*February 17, 2026*

I just built a multi-app platform in an afternoon using Go and HTMX. No React. No Vue. No build step. No node_modules. Just clean server-side code and dynamic UIs that feel modern.

Here's why this stack is criminally underrated for personal projects.

## The JavaScript Fatigue is Real

We all know the drill. You want to build a simple CRUD app. Next thing you know:

```bash
npm install react react-dom
npm install @tanstack/react-query axios
npm install @hookform/resolvers zod
npm install @radix-ui/react-dialog
npm install tailwindcss postcss autoprefixer
npm install -D @types/react @types/node
npm install -D vite @vitejs/plugin-react

# 300MB later...
```

Then you write:
- API route handlers
- React components
- State management
- Type definitions (twice: backend + frontend)
- API client code
- Cache invalidation logic
- Loading states
- Error boundaries
- And so on...

For what? A todo list?

## Enter HTMX + Go

Here's the same todo app:

**Go handler (complete):**
```go
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
    text := r.FormValue("text")
    
    result, _ := h.DB.Exec(
        `INSERT INTO todos (text, username) VALUES (?, ?)`,
        text, auth.GetUsername(r),
    )
    
    id, _ := result.LastInsertId()
    
    fmt.Fprintf(w, `
        <li id="todo-%d">
            <span>%s</span>
            <button hx-delete="/todos/%d" hx-target="#todo-%d" hx-swap="outerHTML">
                Delete
            </button>
        </li>
    `, id, text, id, id)
}
```

**HTML (complete):**
```html
<form hx-post="/todos" hx-target="#todo-list" hx-swap="beforeend">
    <input name="text" placeholder="What needs to be done?" />
    <button type="submit">Add</button>
</form>

<ul id="todo-list">
    <!-- Todos go here -->
</ul>
```

That's it. No build step. No state management. No API client. It just works.

## How HTMX Actually Works

HTMX extends HTML with attributes that make AJAX requests and swap content:

```html
<button 
    hx-post="/items"           <!-- POST to this URL -->
    hx-target="#list"          <!-- Put response here -->
    hx-swap="beforeend"        <!-- Append to end -->
>
    Add Item
</button>
```

When clicked:
1. HTMX sends POST to `/items`
2. Server returns HTML fragment
3. HTMX inserts it into `#list`

No JavaScript written. No `fetch()` calls. No `setState()`. Just declarative attributes.

## Real-World Example: Shopping List

I built a shopping list with multiple stores, items, quantities, and checkboxes. Here's the "add item" flow:

**Form in store card:**
```html
<form hx-post="/stores/{{storeID}}/items" 
      hx-target="#store-{{storeID}}-items" 
      hx-swap="beforeend">
    <input name="name" placeholder="Add item..." required />
    <input name="quantity" placeholder="Qty" />
    <button type="submit">+</button>
</form>

<div id="store-{{storeID}}-items">
    <!-- Items appear here -->
</div>
```

**Handler returns HTML:**
```go
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
    storeID := chi.URLParam(r, "storeID")
    name := r.FormValue("name")
    quantity := r.FormValue("quantity")
    username, _ := auth.GetUsername(r)
    
    result, _ := h.DB.Exec(`
        INSERT INTO items (store_id, name, quantity, username) 
        VALUES (?, ?, ?, ?)
    `, storeID, name, quantity, username)
    
    itemID, _ := result.LastInsertId()
    
    fmt.Fprintf(w, `
        <div class="flex items-center gap-2 p-2 hover:bg-gray-50 rounded">
            <input type="checkbox"
                   hx-post="/items/%d/toggle"
                   hx-target="closest div"
                   hx-swap="outerHTML" />
            <span class="flex-1">%s</span>
            <span class="text-sm text-gray-500">%s</span>
            <button hx-delete="/items/%d"
                    hx-target="closest div"
                    hx-swap="outerHTML">✕</button>
        </div>
    `, itemID, name, quantity, itemID)
}
```

When user submits:
1. Form serializes to `name=Milk&quantity=2L`
2. POST hits Go handler
3. Item saved to SQLite
4. HTML returned
5. HTMX appends it to the list

**No React component. No Redux action. No cache update. Just HTML in, HTML out.**

The checkbox works too! Click it:
1. `hx-post="/items/123/toggle"` fires
2. Server toggles `checked` field
3. Returns updated HTML with `checked` attribute
4. HTMX swaps in place

Everything stays in sync because **the server is the source of truth**.

## The Go Side: Simplicity Wins

Go is perfect for this pattern. It's boring in the best way.

**Server setup (complete):**
```go
func main() {
    db, _ := database.Open("data/app.db")
    defer db.Close()
    
    h := &Handler{DB: db, JWTSecret: "secret"}
    
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    
    // Public routes
    r.Get("/login", h.LoginPage)
    r.Post("/login", h.Login)
    
    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(auth.Middleware(jwtSecret))
        r.Get("/", h.Home)
        r.Post("/items", h.CreateItem)
        r.Delete("/items/{id}", h.DeleteItem)
    })
    
    http.ListenAndServe(":3001", r)
}
```

No complex DI framework. No decorator hell. No "where does this come from?" Just a struct with dependencies and functions that take `http.ResponseWriter`.

**Single binary deployment:**
```bash
CGO_ENABLED=1 go build -o app
./app  # That's it. No runtime. No interpreter.
```

The binary is 17MB. Contains the entire app. Ships to production via `scp`. Runs forever on a $6/month VPS.

Compare that to Next.js:
- Node.js runtime (memory hungry)
- npm dependencies to install
- Environment-specific builds
- PM2 or similar to keep alive
- Potential security updates in 500 packages

## When This Stack Shines

### ✅ Perfect For:

**Personal apps** - You're the only user (or a handful). No need for React's complexity.

**CRUD apps** - Most of what we build: lists, forms, tables. HTMX excels here.

**Internal tools** - Speed matters more than framework choice.

**Rapid prototyping** - No build step = instant feedback loop.

**Learning** - Beginners can see the full request/response cycle.

### ❌ Not Ideal For:

**Highly interactive UIs** - Real-time collaboration, canvas editors, games. Stick to React/Vue.

**Large teams** - If you need TypeScript everywhere and strong contracts, GraphQL might be better.

**Mobile apps** - Though HTMX works on mobile web, native apps need different approaches.

**Heavy client-side logic** - Complex calculations, offline-first. JavaScript frameworks excel here.

## The Techniques That Make It Work

### 1. Target Closest Element

```html
<button hx-delete="/items/123" 
        hx-target="closest div" 
        hx-swap="outerHTML">
    Delete
</button>
```

Deletes the item and removes its container. No ID juggling.

### 2. Return Empty for Deletes

```go
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
    itemID := chi.URLParam(r, "id")
    h.DB.Exec(`DELETE FROM items WHERE id = ?`, itemID)
    w.WriteHeader(http.StatusOK)  // Empty response = remove element
}
```

### 3. Modals Without JavaScript

```html
<button hx-get="/items/new" hx-target="#modal">New Item</button>
<div id="modal"></div>
```

Handler returns modal HTML:
```go
func (h *Handler) NewItemForm(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, `
        <div class="fixed inset-0 bg-black bg-opacity-50">
            <div class="bg-white p-6">
                <form hx-post="/items">
                    <!-- Form fields -->
                </form>
                <button onclick="this.closest('.fixed').remove()">Cancel</button>
            </div>
        </div>
    `)
}
```

### 4. Optimistic UI Updates

```html
<form hx-post="/items" hx-swap="beforeend" hx-target="#list">
    <input name="text" 
           hx-on::after-request="this.value=''" 
           placeholder="Add item" />
</form>
```

Input clears immediately, even though server hasn't responded yet.

### 5. Validation Messages

```go
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
    name := r.FormValue("name")
    
    if len(name) < 3 {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, `<div class="error">Name too short</div>`)
        return
    }
    
    // Create item...
}
```

Return inline error HTML. No complex state management.

## Performance Surprises

"But isn't server rendering slow?" 

Not really. My shopping list app:
- **Initial load:** 120ms (includes auth check, DB queries, HTML gen)
- **Add item:** 15ms (insert + return HTML)
- **Toggle item:** 8ms (update + return HTML)
- **Delete item:** 6ms (delete + empty response)

These are on a $6 Digital Ocean droplet with SQLite.

For comparison, a typical React app:
- **Initial load:** ~2s (bundle download + hydration)
- **Add item:** 100ms (API call + state update + rerender)

The secret? No JavaScript bundle. HTMX is 14KB. My app CSS (Tailwind CDN) loads from cache. The actual data transfer is tiny because HTML compresses well.

Plus, SQLite on the same machine as the app is insanely fast. No network hop to PostgreSQL.

## The Developer Experience

What I love most is the **feedback loop**:

```bash
# Edit handler
vim handlers/items.go

# Build (1 second)
go build

# Run
./app

# Test in browser
# Change something
# Ctrl+C, build, run again
```

Or with Air (Go's hot reload):
```bash
air  # Auto-rebuilds on change
```

No webpack. No node_modules. No waiting for rebuilds. Just code → refresh → see it.

## Real Project Structure

Here's my actual shopping list app:

```
shopping-list/
├── main.go                  # 50 lines
├── handlers/
│   ├── handler.go          # Auth, home page
│   ├── items.go            # CRUD operations
│   └── migrations.go       # DB schema
├── config.yaml             # Users list
└── data/
    └── shopping-list.db    # SQLite DB
```

Total: **~600 lines of Go**. Zero lines of JavaScript (except HTMX from CDN).

Features:
- Multiple stores with custom colors
- Items with quantities
- Check off items as you shop
- Delete stores/items
- User authentication
- Responsive design (Tailwind)

Try building that in React in 600 lines. I'll wait.

## Common Questions

**Q: What about SEO?**
A: These are personal apps behind login. SEO doesn't matter. But if it did, you're server-rendering anyway—perfect for SEO.

**Q: What about accessibility?**
A: HTML forms and buttons are naturally accessible. Screen readers work great. HTMX maintains focus correctly.

**Q: Can I use JavaScript libraries?**
A: Yes! Add Alpine.js for local interactivity:
```html
<div x-data="{ open: false }">
    <button @click="open = !open">Toggle</button>
    <div x-show="open">Content</div>
</div>
```

**Q: What about TypeScript safety?**
A: You lose type checking between frontend/backend. But for personal apps, rapid iteration > type safety. If you need it, generate TypeScript from Go structs.

**Q: Is this production-ready?**
A: I'm running 3 apps on this stack in "production" (just me as user). Zero downtime in 2 months. SQLite handles my load fine.

## The Bottom Line

For personal projects and internal tools, HTMX + Go delivers 80% of React's UX with 20% of the complexity.

You ship faster. You deploy simpler. You understand everything.

Is it the right choice for your startup's main product? Probably not.

Is it perfect for your side project, personal tools, and internal dashboards? Absolutely.

Try it for your next small project. You might be surprised how little you miss React.

---

**Resources:**
- HTMX: [htmx.org](https://htmx.org)
- Go + Chi router: [github.com/go-chi/chi](https://github.com/go-chi/chi)
- My project template: [github.com/yourusername/a-apps](https://github.com)
- HTMX Discord: Great community for questions

*Building something with HTMX? I'd love to see it! [@yourhandle](https://twitter.com)*
