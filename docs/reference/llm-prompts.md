# LLM-Assisted Development Guide

This document provides effective prompts and strategies for using LLM coding assistants (like GitHub Copilot, Claude, ChatGPT) to work with this codebase.

## Repo constraints to remind the LLM about

- **UI contract**: follow [design-spec.md](./design-spec.md)
  - no inline styles
  - use only CSS classes defined in `custom.css`
  - do not introduce Tailwind/Bootstrap/etc.
- Prefer small, explicit changes and reuse patterns from existing apps.

## Why This Structure Works Well with LLMs

1. **Flat, explicit structure** - Easy for LLMs to understand file organization
2. **Minimal abstraction** - Clear, direct code without magic
3. **Reusable patterns** - Reference existing apps as examples
4. **Template-based** - Generate new apps from established patterns
5. **Small scope** - Each app is independent and manageable

## Creating a New App with LLM Assistance

### Step 1: Generate the App Shell

```bash
./scripts/new-app.sh my-new-app 3002
```

### Step 2: Describe Your App to the LLM

Use this prompt template:

```
I have a Go + HTMX app in apps/my-new-app/ that needs to become a [DESCRIPTION].

Reference: apps/todo-list/ shows a complete example with:
- Stores and Items (1-to-many relationship)
- CRUD operations with HTMX
- SQLite database with migrations
- Authentication middleware

Please help me:
1. Design the database schema (tables and relationships)
2. Create migrations in handlers/migrations.go
3. Implement handlers with HTMX patterns
4. Update main.go with the routes

The app should [SPECIFIC REQUIREMENTS].
```

### Step 3: Iterative Development

Ask the LLM to help with specific features:

```
Add a feature to [FEATURE NAME]:
- Database: [table/column changes]
- Handler: [HTMX endpoint description]
- UI: [user interaction]

Use the same HTMX patterns from todo-list (hx-post, hx-delete, hx-target).
```

## Example Prompts

### Creating an Expense Tracker

```
Create an expense tracking app in apps/expense-tracker/ (port 3002).

Reference apps/todo-list/ for structure.

Requirements:
- Track expenses with: date, amount, category, description
- Categories: Food, Transport, Entertainment, Housing, Other
- Monthly view with total
- Filter by category and date range
- Mark expenses as recurring

Database schema:
- categories table (id, name, color, username)
- expenses table (id, category_id, amount, date, description, recurring, username)

HTMX features:
- Add expense form (returns new expense HTML)
- Toggle recurring status
- Delete expense
- Filter form (updates expense list)

Use the same auth patterns and SQLite helpers from shared packages.
```

### Creating a Project Tracker

```
Create a hobby project tracker in apps/project-tracker/ (port 3003).

Reference apps/todo-list/ for HTMX patterns and auth.

Requirements:
- Projects with name, description, status (Not Started, In Progress, Completed, Paused)
- Tasks within each project with checkbox completion
- Progress bar showing completed/total tasks
- Color-coded project cards

Database:
- projects table: id, name, description, status, color, username
- tasks table: id, project_id, name, completed, created_at, username

UI: Grid of project cards similar to todo-list stores
```

### Adding a Feature

```
Add export functionality to todo-list:

1. Add a "Export" button to each store
2. When clicked, download a text file with:
   - Store name
   - List of all unchecked items with quantities
   - Format: "[ ] Item Name (Quantity)"

3. Use Go's text/template or simple string building
4. Handler: GET /stores/{id}/export
5. Set Content-Disposition header for download
```

### Fixing a Bug

```
In todo-list, when I add an item, it appears at the bottom instead of top.

The issue is in handlers/items.go, CreateItem function.
The SQL query sorts by created_at DESC but HTMX uses hx-swap="beforeend".

Fix: Change hx-swap="beforeend" to hx-swap="afterbegin" in the form.
```

## Best Practices with LLMs

### DO:

✅ **Reference existing code** - "Use the same pattern as todo-list"
✅ **Be specific** - "Create a handler at POST /items that returns HTML"
✅ **Mention constraints** - "Use only Go standard library for this"
✅ **Request tests** - "Add a test for the toggle function"
✅ **Ask for complete files** - "Show me the full handler.go file"

### DON'T:

❌ **Be vague** - "Make it better"
❌ **Mix concerns** - "Add auth and also refactor the database"
❌ **Assume magic** - "Automatically persist to DB" (be explicit about where)
❌ **Request framework-specific features** - Stick to Go stdlib + HTMX

## Common Tasks

### Add a new database table

```
Add a "tags" feature to todo-list:

1. Migration in handlers/migrations.go:
   CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT, color TEXT, username TEXT)
   CREATE TABLE item_tags (item_id INTEGER, tag_id INTEGER, FOREIGN KEY...)

2. Update Item struct to include Tags []string

3. Add handlers:
   - GET /tags/new - return modal form
   - POST /tags - create tag
   - POST /items/{id}/tags - attach tag to item

4. Update item HTML to show tags as colored badges
```

### Add authentication to a public endpoint

```
In apps/my-app/main.go, move the /export endpoint from public routes to the protected group:

Current: r.Get("/export", h.Export)  // Outside auth
Change to: Inside r.Group(func(r chi.Router) { ... }) block

This will require login before accessing export.
```

### Style improvements

```
Improve the todo list UI (without changing the visual language):

1. Follow the UI contract in docs/reference/design-spec.md
2. Use only classes from `custom.css` (no new classes, no inline styles)
3. Prefer spacing, `.panel`, and `.card` over new visual variants
4. Avoid icons unless explicitly requested

Point me to the exact files and show the full updated HTML fragments/handlers.
```

## Debugging with LLM

When you encounter errors:

```
I'm getting this error when building todo-list:
[PASTE ERROR]

The error is in [FILE:LINE]. Here's the relevant code:
[PASTE CODE CONTEXT - 10 lines before and after]

What's wrong and how do I fix it?
```

## Advanced: Generate Multiple Apps

```
I want to create 3 apps with similar structure:

1. Books Tracker (port 3004):
   - Track books: title, author, status (Want to Read, Reading, Completed)
   - Notes and rating per book
   - Filter by status

2. Workout Log (port 3005):
   - Log exercises: name, sets, reps, weight, date
   - Group by workout sessions
   - View history by exercise

3. Recipe Manager (port 3006):
   - Save recipes: name, ingredients (list), instructions, tags
   - Search by ingredient or tag
   - Mark favorites

For each:
- Use the scaffold script to create base app
- Reference todo-list for HTMX patterns
- Keep it simple (no search initially, just list/create/delete)
- Use similar color scheme and layout

Start with Books Tracker. Once that works, we'll do the others.
```

## Tips for Effective Prompting

1. **Show, don't tell** - Paste existing code as examples
2. **One feature at a time** - Don't ask for 10 changes at once
3. **Verify incrementally** - Test after each addition
4. **Save working states** - Git commit after each successful feature
5. **Learn patterns** - After LLM shows you once, try to replicate yourself

## Workflow Example

```bash
# 1. Create new app
./scripts/new-app.sh recipe-manager 3006

# 2. Describe to LLM (paste todo-list code as reference)
# 3. LLM provides schema and handlers
# 4. Copy code to files

# 5. Build and test
cd apps/recipe-manager
go build
./recipe-manager  # Test at localhost:3006

# 6. Iterate with LLM for improvements
# 7. Commit when working
git add apps/recipe-manager
git commit -m "Add recipe manager app"
```

## Common Patterns to Reference

Point LLMs to these patterns when needed:

- **CRUD with HTMX**: `apps/todo-list/handlers/items.go`
- **Authentication**: `shared/auth/middleware.go`
- **Database migrations**: `apps/todo-list/handlers/migrations.go`
- **Login flow**: `shared/auth/auth.go` and handlers
- **Base HTML layout**: `shared/templates/base.go`
- **Dockerfile**: `templates/app-template/Dockerfile.tmpl`

## Troubleshooting LLM Responses

### LLM suggests complex solution

Response: "Keep it simple. Use the same pattern as todo-list which just returns HTML."

### LLM uses different framework

Response: "This project uses Go + HTMX only. No React, Vue, or other JS frameworks."

### LLM suggests database changes without migration

Response: "Add the schema change to the Migrations array in handlers/migrations.go."

### LLM provides incomplete code

Response: "Show me the complete [FILE] with all imports and functions, not just the changes."

---

Remember: This structure is designed to be **LLM-friendly**. The simpler and more explicit you make your prompts, the better results you'll get!
