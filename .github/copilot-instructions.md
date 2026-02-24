# Copilot instructions for a-apps

## Big picture (how this repo is shaped)
- Monorepo of small Go + HTMX apps under `apps/<app>/`.
- Each app is self-contained: `main.go`, `handlers/`, `config.yaml`, and a per-app SQLite DB file under `data/`.
- Shared code lives in Go modules under `shared/{auth,database,templates,models}` and is consumed by apps via `replace` directives in each app’s `go.mod`.

## Runtime architecture (request → response)
- Router: `chi` in `apps/<app>/main.go`.
- Auth: login form posts to `/login`, sets a JWT cookie named `auth_token`; protected routes are wrapped by `shared/auth.Middleware`.
- Data: SQLite opened via `shared/database.Open(...)` (foreign keys enabled) and migrations run via `db.RunMigrations(handlers.Migrations)`.
- UI: pages use `shared/templates.BaseHTML` / `LoginHTML`; most app-specific UI is built as HTML strings/fragments returned by handlers (HTMX swaps).
- Shared styling: `GET /custom.css` serves embedded CSS from `shared/templates/custom.css`.

## Developer workflows (the commands people actually use)
- Run an app: `cd apps/<app> && go run .` (uses env `PORT` and `JWT_SECRET`, with dev defaults in code).
- Hot reload: `make install-tools` then `make dev-<app>` (uses `air` if available; falls back to `go run main.go`).
- Tests: `make test` (runs `go test ./...` in each `apps/*` folder), or `cd apps/<app> && go test ./...`.
- Docker (prod-like): `make build`, `make up`, `make down`, `make logs` (uses `deploy/docker-compose.yml` + `deploy/nginx.conf`).

## Repo-specific patterns (copy these)
- User isolation is per-row: tables include a `username` column, and queries filter by it (see `apps/todo-list/handlers/*` and `docs/patterns.md`).
- HTMX handlers commonly return an HTML fragment (not JSON) and rely on `hx-post`/`hx-delete` + `hx-target` + `hx-swap`. For “reload the whole page”, handlers set `HX-Redirect: /`.
- Migrations are plain SQL strings in a `var Migrations = []string{ ... }` slice (see `apps/todo-list/handlers/migrations.go`).

## UI contract (strict)
- Follow `docs/design-spec.md` exactly.
- Use semantic HTML and the shared layout from `shared/templates/base.go`.
- Never use inline styles.
- Only use CSS classes defined in `shared/templates/custom.css` (utilities like `.row`, `.space-between`, `.mt-*`, `.mb-*`, `.panel`, `.primary`, `.muted`, `.center`). Do not invent new classes.

## Config & secrets
- App auth users come from `apps/<app>/config.yaml` (`users:` list with bcrypt hashes).
- Generate password hashes with `./scripts/hash-password.sh <password>`.
- Deploy script uses `DEPLOY_HOST`/`DEPLOY_USER`/`DEPLOY_PATH` (see `scripts/deploy.sh`).

## Common gotcha
- If Go tooling complains about `go.work` entries (apps added/removed), update `go.work` to match `apps/*` you’re working on (or run `go work use ./apps/<app>`).
