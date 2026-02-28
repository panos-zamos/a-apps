# Architecture Refactor Proposal (Base-Path + Local Docker)

## Context / Requirements (from user requests)
- Run the stack locally with Docker Compose and document the workflow.
- Ensure routing works correctly behind the reverse proxy at subpaths (e.g., `/todo`, `/projects`).
- Make behavior consistent for local Docker and DigitalOcean deployments.

## What Was Not Working
1. **Go module replace paths during Docker builds**
   - App Docker builds used a context under `apps/*`, but `go.mod` uses `replace` directives to `../../shared/*`.
   - Inside the container, `/shared` did not exist, so `go mod download` failed.

2. **CGO build failure for sqlite3**
   - App Dockerfiles built with `CGO_ENABLED=1`, but Alpine builder images lacked a C compiler.
   - Result: `cgo: C compiler "gcc" not found`.

3. **Reverse-proxy path routing**
   - Caddy routed `/todo` to the app without stripping the prefix, so the app saw `/todo` and returned 404.
   - Redirects and absolute links pointed to `/login`, `/custom.css`, and `/`, breaking when the app lived under `/todo` or `/projects`.

4. **Inconsistent documentation**
   - README referenced nginx and old compose paths; deploy layout uses Caddy and `deploy/docker-compose.yml`.

## High-Level Fixes Applied (Code + Config)
- Use repo-root build context in Docker Compose and copy `shared/` into builder images.
- Install `build-base` in builder stage so CGO builds succeed.
- Add a `BASE_PATH` runtime config and prefix all redirects/links/HTMX URLs with it.
- Update Caddy to redirect `/todo` → `/todo/` (and `/projects` → `/projects/`) so `handle_path` can strip the prefix.
- Document local compose usage and Caddy-based deployment.

---

## Proposal: Architecture Refactor (Doc-Only)

### Goals
- Make subpath hosting a first-class concern.
- Reduce future regressions from hard-coded absolute URLs.
- Separate application routing from reverse-proxy routing in a consistent and testable way.
- Avoid Docker build failures from local module replacements.

### Proposal A (Recommended): Standardize Base-Path Support
**Summary:** Introduce a shared URL helper package and make `BASE_PATH` a first-class config for every app.

**Key ideas:**
- Add a small shared helper (e.g., `shared/web/path.go`) to generate absolute and asset URLs.
- Update templates and handlers to use these helpers.
- Add `BASE_PATH` into `config.yaml` and environment override logic for every app.
- Add tests for URL generation so `/todo/login` and `/projects/custom.css` work consistently.

**Pros:**
- Eliminates hard-coded `"/"` paths in templates and handlers.
- Keeps reverse-proxy config simple and consistent.
- Better for hosting multiple apps under one domain.

**Cons:**
- Requires touching every app/template when added.

### Proposal B: Internal Router with Subrouter Prefix
**Summary:** Instead of transforming URLs, mount app routers under a configurable prefix.

**Key ideas:**
- Define `basePath` and mount a subrouter with `r.Mount(basePath, appRouter)`.
- Use route generation to create relative URLs for templates.

**Pros:**
- Cleaner separation between prefix routing and app routes.

**Cons:**
- Requires middleware and URL helpers anyway for absolute links and HTMX headers.
- Harder to reason about when the app is run directly on `localhost:3001` without a proxy.

### Proposal C: Split reverse-proxy responsibilities
**Summary:** Keep apps on root paths and use subdomain routing (e.g., `todo.example.com`).

**Pros:**
- Avoids base-path complexity entirely.

**Cons:**
- Requires DNS + TLS setup per app, not ideal for a single droplet with minimal overhead.

### Recommendation
Adopt **Proposal A**. It is the least disruptive to local development while solving subpath hosting robustly. It also aligns with how HTMX apps need explicit, predictable URL generation.

---

## Suggested Follow-ups
- Create `shared/web` helpers (link/path builder, HTMX redirect helper).
- Add `BasePath` to template contexts in all apps.
- Add lint/grep check that flags hard-coded `"/"` paths in templates.
- Add a small integration test that runs an app with `BASE_PATH=/todo` and verifies redirects.

## Affected Docs (if implemented later)
- README (local usage + proxy behavior)
- docs/reference/patterns.md (URL helper patterns)
- docs/blog/blog-zero-to-deployed.md (deployment details)
