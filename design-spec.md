Here is the content for `design-spec.md`:

---

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

* Light mode only
* Neutral grayscale palette
* Soft rounded corners
* Subtle borders instead of shadows
* Large readable typography
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
    <aside class="sidebar">
      <!-- navigation -->
    </aside>

    <main class="content">
      <div class="content-inner">
        <!-- page content -->
      </div>
    </main>
  </div>
</body>
```

## Layout Rules

* Sidebar width: 240px
* Sidebar contains navigation only
* Main content scrolls
* Content width limited using `.content-inner`
* Maintain vertical spacing rhythm (8px / 16px / 24px)

Do not modify layout structure unless explicitly instructed.

---

# 3. Typography

## Font

* Use system font stack (already defined in CSS)
* Base font size: 16px

## Headings

* `h1` for page title (one per page)
* `h2` for section titles
* Avoid skipping heading levels

## Text Usage

* Use `.muted` for metadata or secondary text
* Do not reduce font size manually
* Do not use bold excessively

Hierarchy should come from structure, not styling tricks.

---

# 4. UI Elements

## Panels

Use `.panel` for grouped content:

```html
<div class="panel">
  <h2>Section</h2>
  <p>Content</p>
</div>
```

Rules:

* Panels contain logical groups
* Do not nest panels unnecessarily

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

Rules:

* Only one `.primary` button per action group
* Do not create visual variants
* Do not change button sizes

---

## Forms

Inputs are full-width by default.

Example:

```html
<form>
  <label>Name</label>
  <input type="text" name="name" required>

  <div class="mt-md">
    <button class="primary">Submit</button>
  </div>
</form>
```

Rules:

* Use proper labels
* Do not wrap inputs in custom containers
* Use spacing utilities instead of inline margins

---

## Layout Utilities (Allowed)

Agents may use only these utility classes:

* `.row`
* `.space-between`
* `.mt-sm` / `.mt-md` / `.mt-lg`
* `.mb-sm` / `.mb-md` / `.mb-lg`
* `.center`
* `.panel`
* `.primary`
* `.muted`

No other classes are allowed.

---

# Final Advice for Agents

1. Do not invent new components.
2. Do not add decorative elements.
3. Do not override CSS.
4. Do not inline styles.
5. Do not introduce new layout patterns.
6. Prefer simple vertical stacking.
7. Use `.panel` for grouping.
8. Use `.row` only when horizontal alignment is necessary.
9. Use `.primary` only for the main action.

If uncertain, choose the simpler layout.

Constraint ensures long-term consistency.
