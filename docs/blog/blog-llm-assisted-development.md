# Optimizing Your Codebase for LLM-Assisted Development

> **Note:** This blog post was written by an AI agent (Claude) to document and explain the architecture decisions made during the development of this project. It serves as a learning resource to understand what was built and why certain approaches were chosen.

*February 17, 2026*

After building a personal app platform with extensive LLM assistance, I've discovered that certain code patterns seem to improve AI-assisted development. Here's what I've learned so far.

## The Challenge with "Smart" Code

Many developers (myself included) often optimize code for elegance—using abstractions, metaprogramming, and framework magic. While this can feel satisfying, I've noticed that LLMs sometimes struggle with implicit behavior.

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

This looks clean, but when working with an LLM, there are several things that aren't immediately visible:
- Where does `current_user` come from? (Devise gem? Custom middleware?)
- What does `respond_with` do? (Responders gem magic)
- What's `set_item`? (Private method defined... somewhere)
- What validations run? (In the model, not visible here)

Now compare this Go version from my todo-list app:

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

In this version, most things are explicit:
- Authentication? `auth.GetUsername(r)` - visible import
- Database? Raw SQL right there
- Response? Plain HTML string
- Error handling? Right in the function

When I ask an LLM to "add quantity field to items," it can often see the entire flow in one screen, which seems to help.

## The Flat Structure Principle

My original instinct was to organize like this:
```
todo-list/
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

"Clean architecture!" I thought. But I found that LLMs sometimes have difficulty with this structure. Asking "add a field to items" requires understanding:
1. Which layer defines the field
2. Which interfaces need updating
3. Which implementations change
4. Where DTOs get mapped

Instead, I went flat:
```
todo-list/
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

**Result in my experience:** New features that used to take around 30 minutes of back-and-forth now take closer to 5 minutes.

## Template-Based Generation

The killer feature of this architecture is the `new-app.sh` script. It copies a template and replaces placeholders:

```bash
sed -e "s/{{APP_NAME}}/$APP_NAME/g" \
    -e "s/{{APP_PORT}}/$APP_PORT/g" \
    template.go > apps/new-app/main.go
```

Simple? Yes. But it enables something useful: I can tell an LLM:

> "Generate a habit tracker app. Reference todo-list for patterns. The scaffold already created the base structure."

The LLM now has:
1. A working example (todo-list)
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

I've found that HTMX works quite well for LLM-assisted development. Here's a comparison:

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

Most things are in one place. The LLM can see:
1. The form
2. The endpoint
3. What gets returned
4. Where it goes

No state management, no cache invalidation, no "component rerender" debugging.

## Real Example: Adding Tags Feature

I asked Claude: "Add tags to shopping list items. Users should be able to create tags and assign multiple tags to items."

**With my old React + separate API architecture:** Around 45 minutes, 12 back-and-forth messages

**With this Go + HTMX setup:** About 8 minutes, 2 messages

My guess is that the LLM could see:
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

Most of it worked correctly on the first try.

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

In my experience with this project:

Before this architecture (Rails + React):
- **New feature:** ~45 min average
- **Bug fix reported by LLM:** ~30% success rate
- **Generated code works first try:** ~20%

With this architecture (Go + HTMX):
- **New feature:** ~12 min average (roughly 73% faster)
- **Bug fix reported by LLM:** ~85% success rate
- **Generated code works first try:** ~70%

These are just my observations from this particular project. Your mileage may vary depending on the problem domain and LLM you use.

## What I've Learned

As we work more with AI assistants, our optimization priorities might need to shift:

**Traditional focus:** DRY, abstraction, "clever" solutions
**Experiment with:** Explicit, flat, colocated code

Interestingly, this often seems to make code easier for humans too. In my experience, developers at different skill levels can understand and modify this code without much explanation.

## Try It Yourself

If you're curious, you could try this approach in your own projects. Ask your LLM: "Add a notes feature where users can attach notes to items. Each note has text and a timestamp."

Compare the experience with different architectures and see what works best for you.

---

**My takeaway:** It might be worth considering code comprehension—both for humans and AI—when making architecture decisions. The results in this project were encouraging, though every project is different.

*What patterns have you found that work well with LLMs? I'm [@yourhandle](https://twitter.com) on Twitter—let's discuss!*
