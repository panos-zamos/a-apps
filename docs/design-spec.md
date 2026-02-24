# Design Spec – AI Agent UI Contract

This document defines strict UI rules for this Go + HTMX project.
Agents must follow this specification precisely.

The goal is:

* Predictable output
* Minimal styling entropy
* Clean, readable interface
* No decorative experimentation

Do not invent new styles. Do not improvise visually.

---

# 1. Overall Principles

## Visual Direction

* Mobile-first, single-column layout
* Light mode only
* Warm neutral palette (`#f5f5f0` background, `#1a1a1a` text)
* Soft rounded corners (12px cards, 8px inputs)
* Subtle borders instead of shadows
* Large readable typography (15px base)
* Clean spacing
* No visual noise

## Implementation Rules

Agents must:

* Prefer semantic HTML (`section`, `article`, `nav`, `header`, `main`)
* Never use inline styles
* Use only classes defined in `custom.css`
* Avoid decorative elements
* Prefer spacing over styling
* Avoid icons unless explicitly required
* Never invent new CSS classes

Consistency > creativity.

---

# 2. Layout Structure

## Page Shell

All pages must follow this structure:

```html
<body>
  <div class="app">
    <header class="top-bar">
      <h1>App Name</h1>
      <!-- optional: logout form -->
    </header>

    <main class="content">
      <!-- page content -->
    </main>
  </div>
</body>
```

## Layout Rules

* No sidebar — mobile-first single column
* `.app` is the page container (max-width: 600px, centered)
* `.top-bar` is sticky at the top with app name and controls
* `.content` has 16px horizontal padding
* Maintain vertical spacing rhythm (8px / 16px / 24px)

Do not modify layout structure unless explicitly instructed.

---

# 3. Typography

## Font

* Use system font stack (already defined in CSS: `-apple-system, BlinkMacSystemFont, "Inter", "Segoe UI", system-ui, sans-serif`)
* Base font size: 15px

## Headings

* `h1` for the app name in the top bar (one per page)
* `h2` for section titles within content
* Avoid skipping heading levels

## Text Usage

* Use `.muted` for metadata or secondary text
* Do not reduce font size manually
* Do not use bold excessively

Hierarchy should come from structure, not styling tricks.

---

# 4. UI Elements

## Cards

Use `.card` for clickable content items in lists:

```html
<a href="/item/1" class="card">
  <div class="card-header">
    <span class="card-name">Item Name</span>
    <span class="badge badge-dev">development</span>
  </div>
  <p class="card-desc">Short description</p>
  <div class="card-meta">
    <span class="stars">★★★★☆</span>
    <span class="tags">tag1 · tag2</span>
  </div>
</a>
```

Use `.card-list` to wrap a vertical list of cards with consistent gap:

```html
<div class="card-list">
  <a href="..." class="card">...</a>
  <a href="..." class="card">...</a>
</div>
```

Cards can also be used as non-clickable containers (using `<article class="card">`).

---

## Panels

Use `.panel` for grouped content that is not a list item:

```html
<div class="panel">
  <h2>Section</h2>
  <p>Content</p>
</div>
```

Rules:

* Panels contain logical groups (forms, configuration)
* Do not nest panels unnecessarily

---

## Badges

Use `.badge` with a stage modifier for status indicators:

```html
<span class="badge badge-dev">development</span>
```

Available badge modifiers: `.badge-idea`, `.badge-planning`, `.badge-dev`, `.badge-released`, `.badge-archived`

---

## Buttons

Primary action:

```html
<button class="primary">Save</button>
```

Secondary action:

```html
<button>Cancel</button>
```

Destructive action:

```html
<button class="danger">Delete</button>
```

Circular add button:

```html
<button class="btn-add">+</button>
```

Rules:

* Only one `.primary` button per action group
* Use `.danger` for destructive actions
* Use `.btn-add` for the main "add/create" action on list pages
* Do not create other visual variants
* Do not change button sizes

---

## Filters

Use `.filters` for horizontally scrolling filter controls:

```html
<div class="filters mb-md">
  <select name="filter1">...</select>
  <select name="filter2">...</select>
</div>
```

Filters scroll horizontally on small screens.

---

## Detail Page Meta

Use `.detail-meta` with `.meta-row` for key-value metadata:

```html
<div class="detail-meta">
  <div class="meta-row">
    <span class="meta-label">Label</span>
    <span>Value</span>
  </div>
</div>
```

---

## Timeline / Log Entries

Use `.timeline` with `.log-entry` for chronological entries:

```html
<div class="timeline">
  <article class="log-entry">
    <div class="log-date">2026-02-24</div>
    <div class="log-note">Entry text</div>
    <div class="log-actions">
      <button>Reply</button>
      <button>Delete</button>
    </div>
  </article>
  <article class="log-entry nested">
    <div class="log-date">2026-02-24</div>
    <div class="log-note">Nested reply</div>
  </article>
</div>
```

`.nested` indents child entries and adds a left border.

---

## Section Titles

Use `.section-title` for uppercase label-style headings:

```html
<div class="section-title">Description</div>
```

---

## Navigation

Use `.back-link` for back navigation:

```html
<a href="/" class="back-link">&larr; Back</a>
```

---

## Forms

Inputs are full-width by default.

Example:

```html
<form>
  <label>Name</label>
  <input type="text" name="name" required>

  <div class="mt-md">
    <label>Description</label>
    <textarea name="description"></textarea>
  </div>

  <div class="mt-md">
    <button class="primary">Submit</button>
  </div>
</form>
```

For checkbox groups, use `.checkbox-row`:

```html
<div class="checkbox-row">
  <label><input type="checkbox" name="opt1"> Option 1</label>
  <label><input type="checkbox" name="opt2"> Option 2</label>
</div>
```

Rules:

* Use proper labels
* Do not wrap inputs in custom containers
* Use spacing utilities instead of inline margins

---

## Layout Utilities (Allowed)

Agents may use only these utility classes:

* `.row` — flex row with 12px gap
* `.space-between` — justify space-between
* `.mt-sm` / `.mt-md` / `.mt-lg` — top margin (8/16/24px)
* `.mb-sm` / `.mb-md` / `.mb-lg` — bottom margin (8/16/24px)
* `.center` — centered text
* `.panel` — bordered content group
* `.card` / `.card-list` / `.card-header` / `.card-name` / `.card-desc` / `.card-meta` — card components
* `.badge` + stage modifiers — status badges
* `.primary` — primary action
* `.danger` — destructive action
* `.btn-add` — circular add button
* `.muted` — muted text
* `.stars` / `.tags` — display helpers
* `.filters` — filter row
* `.back-link` — back navigation
* `.detail-meta` / `.meta-row` / `.meta-label` / `.meta-value` — detail metadata
* `.section-title` — uppercase section label
* `.timeline` / `.log-entry` / `.nested` / `.log-date` / `.log-note` / `.log-actions` — timeline
* `.checkbox-row` — checkbox group

No other classes are allowed.

---

# Final Advice for Agents

1. Do not invent new components.
2. Do not add decorative elements.
3. Do not override CSS.
4. Do not inline styles.
5. Do not introduce new layout patterns.
6. Prefer simple vertical stacking.
7. Use `.card` for list items, `.panel` for forms/config.
8. Use `.row` only when horizontal alignment is necessary.
9. Use `.primary` only for the main action.
10. Design for mobile first — 600px max width, touch-friendly targets.

If uncertain, choose the simpler layout.

Constraint ensures long-term consistency.
