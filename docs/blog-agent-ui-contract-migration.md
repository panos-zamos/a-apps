# UI Contract Migration: Removing Tailwind, Centralizing Styles, and Serving `custom.css`

> Disclaimer: this post was written by an automated coding agent based on the repository changes it performed.

## Context

This repository uses a small UI contract (`docs/design-spec.md`) that enforces:

- light mode, neutral grayscale
- semantic HTML
- no inline styles
- **only** classes defined in `custom.css`
- minimal components: `.app`, `.sidebar`, `.content`, `.content-inner`, `.panel`, buttons, form controls, and a small allowed utility set

The `shopping-list` app previously relied heavily on Tailwind utility classes embedded directly in HTML strings.

## What changed

### 1) Shared templates switched to the contract

File updated:
- `shared/templates/base.go`

Changes:
- Removed Tailwind CDN usage.
- Implemented the required page shell:
  - `.app` + `.sidebar` + `.content` + `.content-inner`
- Linked the shared stylesheet via `<link rel="stylesheet" href="/custom.css">`.
- Simplified login layout to use `.panel`, `.muted`, and spacing utilities.

Result: all apps that use `sharedTemplates.BaseHTML` / `sharedTemplates.LoginHTML` now render with the contract layout by default.

### 2) Shopping-list markup was rewritten to use only allowed classes

Files updated:
- `apps/shopping-list/handlers/handler.go`
- `apps/shopping-list/handlers/items.go`

Changes:
- Replaced Tailwind class soup (e.g. `flex`, `grid`, `text-gray-*`, `shadow`, etc.) with the allowed class set:
  - layout/grouping via `.panel`
  - alignment via `.row` and `.space-between`
  - spacing via `.mt-*` / `.mb-*`
  - secondary text via `.muted`
- Removed inline styles (e.g. `style="border-color: ..."`) and decorative icons.
- Adjusted HTMX `hx-target` selectors to consistently replace the full item row.

### 3) `custom.css` was moved to a stable location and embedded

Motivation: running `go run .` didn’t serve `custom.css`, which meant the new templates rendered unstyled.

Final solution:
- Moved stylesheet to the module that is already in `go.work` and imported by apps:
  - `shared/templates/custom.css`
- Added an embedded asset handler:
  - `shared/templates/custom_css.go`

This uses `go:embed` to package the CSS into the binary and serves it with the proper `Content-Type`.

### 4) Apps now serve `/custom.css` from the binary

Files updated:
- `apps/shopping-list/main.go`
- `templates/app-template/main.go.tmpl`

Change:
- Added a public route:

```go
r.Get("/custom.css", sharedTemplates.CustomCSSHandler())
```

This removes any dependency on the current working directory or Docker copy steps.

### 5) Design spec moved under `docs/`

- `design-spec.md` was relocated to `docs/design-spec.md`.
- A small header/preamble artifact was removed so the document is clean Markdown.

### 6) App template updated to match the contract

Files updated:
- `templates/app-template/handlers/handler.go.tmpl`
- `templates/app-template/main.go.tmpl`

Changes:
- Replaced Tailwind-based placeholder HTML with `.panel` + `.muted`.
- Ensured newly generated apps also serve `/custom.css` via the embedded handler.

## How to verify

From `apps/shopping-list`:

- Start the server:
  - `go run .`
- Open:
  - `http://localhost:3001/login`
  - `http://localhost:3001/`
- Confirm:
  - `/custom.css` returns CSS
  - pages render with sidebar + content shell
  - no Tailwind dependency

Tests run during the change:

- `go test ./apps/shopping-list/...`
- `go test ./shared/templates/...`

## Notes / follow-ups

- The design contract is intentionally strict. If you need additional UI patterns, it’s better to extend `custom.css` deliberately (and update `docs/design-spec.md`) than to add one-off classes.
