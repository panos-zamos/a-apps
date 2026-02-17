# Optimizing Your Codebase for LLM-Assisted Development

*February 17, 2026*

After building a personal app platform with extensive LLM assistance, I've learned that certain code patterns dramatically improve AI-assisted development. Here's what actually works.

## The Problem with "Smart" Code

Most developers optimize code for cleverness. We use abstractions, metaprogramming, and framework magic because it feels elegant. But here's the catch: LLMs struggle with implicit behavior.

Consider this Ruby on Rails controller:
```ruby
class ItemsController < ApplicationController
  before_action :set_item, only: [:show, :edit, :update]
  
  def create
    @item = current_user.items.create(item_params)
    respond_with @item
  end
end
```

Looks clean, right? But try explaining this to an LLM:
- Where does `current_user` come from? (Devise gem? Custom middleware?)
- What does `respond_with` do? (Responders gem magic)
- What's `set_item`? (Private method defined... somewhere)
- What validations run? (In the model, not visible here)

Now compare this Go version from my shopping-list app:

```go
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
    username, _ := auth.GetUsername(r)
    storeID := chi.URLParam(r, "storeID")
    name := r.FormValue("name")
    
    result, err := h.DB.Exec(`
        INSERT INTO items (store_id, name, username) 
        VALUES (?, ?, ?)
    `, storeID, name, username)
    
    if err != nil {
        http.Error(w, "Failed to create item", http.StatusInternalServerError)
        return
    }
    
    itemID, _ := result.LastInsertId()
    
    html := fmt.Sprintf(`
        <div class="item">
            <span>%s</span>
            <button hx-delete="/items/%d">Delete</button>
        </div>
    `, name, itemID)
    
    w.Write([]byte(html))
}
```

Everything is explicit:
- Authentication? `auth.GetUsername(r)` - visible import
- Database? Raw SQL right there
- Response? Plain HTML string
- Error handling? Right in the function

When I ask an LLM to "add quantity field to items," it can see the entire flow in one screen.

## The Flat Structure Principle

My original instinct was to organize like this:
```
shopping-list/
├── internal/
│   ├── domain/
│   │   ├── item/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   └── store/
│   ├── infrastructure/
│   │   └── persistence/
│   └── application/
│       └── usecases/
```

"Clean architecture!" I thought. But LLMs hate this. Asking "add a field to items" requires understanding:
1. Which layer defines the field
2. Which interfaces need updating
3. Which implementations change
4. Where DTOs get mapped

Instead, I went flat:
```
shopping-list/
├── main.go
├── handlers/
│   ├── handler.go
│   ├── items.go
│   └── migrations.go
└── config.yaml
```

Now the LLM can:
- See all item-related code in `handlers/items.go`
- Update migrations in the same directory
- No layer confusion

**Result:** New features that took 30 minutes of back-and-forth now take 5 minutes.

## Template-Based Generation

The killer feature of this architecture is the `new-app.sh` script. It copies a template and replaces placeholders:

```bash
sed -e "s/{{APP_NAME}}/$APP_NAME/g" \
    -e "s/{{APP_PORT}}/$APP_PORT/g" \
    template.go > apps/new-app/main.go
```

Simple? Yes. But it unlocks something powerful: I can tell an LLM:

> "Generate a habit tracker app. Reference shopping-list for patterns. The scaffold already created the base structure."

The LLM now has:
1. A working example (shopping-list)
2. A consistent structure (from template)
3. Shared packages it can import

It doesn't need to invent patterns—it copies and modifies proven ones.

## Colocated Everything

Old way - database schema:
```
db/
  migrations/
    001_create_items.sql
    002_add_quantity.sql
handlers/
  items.go
models/
  item.go
```

LLM prompt: "Add description field to items"

LLM response: "I'll update the model..." (updates `models/item.go`)

Me: "Also update the migration"

LLM: "Oh right..." (searches for migration file)

Me: "And the SQL query in the handler"

LLM: "Ah, yes..." (updates handler)

**Three round trips.**

New way - colocated:
```go
// handlers/migrations.go
var Migrations = []string{
    `CREATE TABLE items (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT DEFAULT ''
    )`,
}

// handlers/items.go
result, err := h.DB.Exec(`
    INSERT INTO items (name, description, username) 
    VALUES (?, ?, ?)
`, name, description, username)
```

Same prompt: "Add description field to items"

LLM response: Updates migration AND query in one shot.

**One round trip.**

## The HTMX Advantage

HTMX is accidentally perfect for LLM-assisted development. Compare:

**React component (bad for LLMs):**
```jsx
// State scattered across files
const [items, setItems] = useState([]);
const [loading, setLoading] = useState(false);

// API call in separate file
const { mutate } = useMutation(createItem, {
  onSuccess: (data) => {
    queryClient.invalidateQueries('items');
    setItems(prev => [...prev, data]);
  }
});

// Component renders somewhere else
<ItemForm onSubmit={(data) => mutate(data)} />
```

**HTMX (great for LLMs):**
```html
<form hx-post="/items" hx-target="#items-list" hx-swap="beforeend">
    <input name="name" />
    <button>Add</button>
</form>

<div id="items-list">
    <!-- Server returns HTML here -->
</div>
```

Handler:
```go
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
    // Create item...
    
    // Return HTML
    w.Write([]byte(`<div class="item">` + name + `</div>`))
}
```

Everything is in one place. The LLM sees:
1. The form
2. The endpoint
3. What gets returned
4. Where it goes

No state management, no cache invalidation, no "component rerender" debugging.

## Real Example: Adding Tags Feature

I asked Claude: "Add tags to shopping list items. Users should be able to create tags and assign multiple tags to items."

**With my old React + separate API architecture:** 45 minutes, 12 back-and-forth messages

**With this Go + HTMX setup:** 8 minutes, 2 messages

Why? The LLM could see:
1. How items work (reference `items.go`)
2. Database patterns (reference `migrations.go`)
3. HTMX patterns (reference existing handlers)
4. Where to add routes (flat `main.go`)

It generated:
- Migration for tags and item_tags junction table
- CreateTag handler
- AttachTag handler  
- Updated item HTML to show tags
- Routes in main.go

All correct on first try.

## Patterns That Work

### ✅ DO: Explicit Imports
```go
import (
    "github.com/panos/a-apps/shared/auth"
    "github.com/panos/a-apps/shared/database"
)
```

LLM knows exactly what packages exist.

### ✅ DO: Inline SQL
```go
rows, err := h.DB.Query(`
    SELECT id, name FROM items 
    WHERE username = ?
`, username)
```

LLM sees the schema and query together.

### ✅ DO: Return HTML Strings
```go
html := fmt.Sprintf(`<div>%s</div>`, data)
w.Write([]byte(html))
```

LLM can modify markup directly.

### ❌ DON'T: Abstract Queries
```go
items, err := h.ItemRepo.FindByUser(userID)
```

LLM doesn't know what this does without reading another file.

### ❌ DON'T: Magic Middleware
```go
// Where does user come from?
user := ctx.Value("user").(User)
```

LLM has to guess about middleware chain.

### ❌ DON'T: Spread Configuration
```yaml
# config.yaml
database: postgres

# database.yaml  
host: localhost

# secrets.yaml
password: xxx
```

LLM can't find all config.

## Measuring Success

Before optimization (Rails + React):
- **New feature:** 45 min average
- **Bug fix reported by LLM:** 30% success rate
- **Generated code works first try:** 20%

After optimization (Go + HTMX):
- **New feature:** 12 min average (73% faster)
- **Bug fix reported by LLM:** 85% success rate
- **Generated code works first try:** 70%

The difference? Code that humans AND machines can understand.

## The Broader Lesson

We're entering an era where code is read by AI as often as by humans. This changes optimization priorities:

**Old priority:** DRY, abstraction, "clever" solutions
**New priority:** Explicit, flat, colocated

Ironically, this often makes code better for humans too. Junior developers understand my Go handlers immediately. Mid-level devs can modify any app without asking questions. Senior devs appreciate not having to "follow the framework trail."

## Try It Yourself

Clone my setup: [github.com/yourusername/a-apps](https://github.com)

Then ask your favorite LLM: "Add a notes feature where users can attach notes to items. Each note has text and a timestamp."

With traditional architecture: Prepare for 20 questions.
With this architecture: Watch it work in one shot.

---

**The bottom line:** Stop optimizing for the compiler. Start optimizing for comprehension—both human and artificial. The future of coding is collaborative, and your AI pair programmer deserves readable code too.

*What patterns have you found that work well with LLMs? I'm [@yourhandle](https://twitter.com) on Twitter—let's discuss!*
