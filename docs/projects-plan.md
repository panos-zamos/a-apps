# `projects` — App Design Plan

Application for planning and tracking projects.

---

## Data Model

### Project

| Field               | Type              | Notes                                                    |
|---------------------|-------------------|----------------------------------------------------------|
| `id`                | integer PK        | auto-increment                                           |
| `username`          | text              | user isolation (repo pattern)                            |
| `short_name`        | text, required    | project identifier                                       |
| `short_description` | text              | ~10 word tagline                                         |
| `full_description`  | text              | detailed description (includes goals)                    |
| `website_url`       | text              | link to the project website                              |
| `source_url`        | text              | link to the source code (e.g. GitHub repo)               |
| `is_commercial`     | boolean           | flag                                                     |
| `is_open_source`    | boolean           | flag                                                     |
| `is_public`         | boolean           | flag                                                     |
| `stage`             | text              | one of: idea, planning, development, released, archived  |
| `rating`            | integer           | 1–5 success/satisfaction scale                           |
| `created_at`        | datetime          | auto                                                     |
| `updated_at`        | datetime          | auto                                                     |

### Log Entry (timeline / research)

| Field        | Type                  | Notes                                  |
|--------------|-----------------------|----------------------------------------|
| `id`         | integer PK            | auto-increment                         |
| `project_id` | integer FK            | belongs to a project                   |
| `parent_id`  | integer FK, nullable  | self-referencing for nesting           |
| `username`   | text                  | user isolation                         |
| `note`       | text                  | the log content                        |
| `url`        | text, nullable        | optional link                          |
| `created_at` | datetime              | auto                                   |

Entries are nested: a log entry can have child entries via `parent_id`, allowing threaded notes under a parent item (e.g. a research link with sub-notes beneath it).

---

## Stages (progression)

```
idea → planning → development → released → archived
```

Stage **is** the progress indicator — no separate progress field.

---

## Pages / Routes

| Route                              | Method   | What it does                        |
|------------------------------------|----------|-------------------------------------|
| `/`                                | GET      | Home — projects table               |
| `/login`                           | GET/POST | Auth (standard)                     |
| `/projects`                        | POST     | Create project                      |
| `/projects/new`                    | GET      | New project form (HTMX)            |
| `/projects/{id}`                   | GET      | Project detail + log timeline       |
| `/projects/{id}`                   | PUT      | Update project                      |
| `/projects/{id}`                   | DELETE   | Delete project                      |
| `/projects/{id}/log`               | POST     | Add log entry                       |
| `/projects/{id}/log/{logId}`       | DELETE   | Remove log entry                    |
| `/projects/{id}/log/{logId}/reply` | POST     | Add nested child entry              |

---

## Home Page (table / list view)

```
┌─────────────────────────────────────────────────────────────┐
│ Projects                                    [+ New Project] │
├─────────────────────────────────────────────────────────────┤
│ Filters: [Stage ▾] [Type ▾] [Rating ▾]                     │
├──────────┬─────────────┬───────┬────────┬───────────────────┤
│ Name     │ Description │ Stage │ Rating │ Flags             │
├──────────┼─────────────┼───────┼────────┼───────────────────┤
│ a-apps   │ Personal... │ dev   │ ★★★★☆ │ open-source       │
│ acme-crm │ Client...   │ idea  │ —      │ commercial        │
└──────────┴─────────────┴───────┴────────┴───────────────────┘
```

- Clicking a row navigates to the project detail page
- Filters use HTMX to reload the table body

---

## Project Detail Page

```
┌─────────────────────────────────────────────────┐
│ ← Back                                          │
│                                                  │
│ a-apps                          [Edit] [Delete]  │
│ Personal app monorepo                            │
│ Stage: [development ▾]  Rating: ★★★★☆            │
│ 🌐 myproject.com                                 │
│ 📦 github.com/user/repo                          │
│ open-source · public                             │
│                                                  │
│ Full description text here...                    │
│                                                  │
├─────────────────────────────────────────────────┤
│ Timeline                           [+ Add Entry] │
│                                                  │
│ 2026-02-24  Set up CI pipeline                   │
│   └─ 2026-02-24  Added GitHub Actions config     │
│   └─ 2026-02-24  Fixed lint step                 │
│                                                  │
│ 2026-02-20  Initial research                     │
│   └─ 2026-02-20  Found useful lib → [link]       │
│                                                  │
│ 2026-02-18  Had the idea                         │
└─────────────────────────────────────────────────┘
```

- Stage can be changed inline (dropdown, HTMX PUT)
- Log entries show newest-first, nested children indented
- "Reply" link on each entry to add a child note

---

## Key Design Decisions

1. **No separate goal/progress fields** — stage is the progress, goals go in the description
2. **Rating is 1–5** — assess success/satisfaction at any point
3. **Two URL fields** — `website_url` for the live site, `source_url` for the repo
4. **Nested log entries** — flat list with optional parent, gives simple threading without over-engineering
5. **Filtering** — by stage, flags (commercial/open-source/public), and rating on the home table
6. **Follows repo conventions** — user isolation via `username`, HTMX fragments, SQLite, shared templates
