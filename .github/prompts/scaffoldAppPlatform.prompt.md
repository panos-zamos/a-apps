---
name: scaffoldAppPlatform
description: Scaffold a multi-app web platform with isolated apps and easy deployments.
argument-hint: Preferred language/stack, deployment target, auth needs, example apps list
---
You are a senior software engineer and DevOps-minded architect.

Goal: Build a collection of small web applications (e.g., shopping list, expense tracker, hobby project tracker) as a cohesive platform with **minimal friction to add new apps**, **strong isolation** (changing one app shouldn’t break others), and **easy deployment** (from laptop or GitHub) to a VPS/PaaS (e.g., DigitalOcean). Optimize the repo for **LLM-agent-assisted development**.

Inputs (fill in or infer):
- Target host: {VPS/DigitalOcean Droplet | App Platform | other}
- Preferred backend language: {Go | PHP | Python | mixed}
- UI approach: {server-rendered + HTMX | SPA | simple MPA}
- Primary user model: {single-user by default | occasional multi-user via config}
- Database preference: {SQLite per app | Postgres shared | managed DB}
- Example apps: {list of 3–10 apps}
- Budget constraints and expected traffic: {low/personal | moderate}

Deliverables:
1) A recommended architecture with trade-offs (monorepo vs multi-repo; shared vs per-app deploy; DB strategy).
2) A concrete repo layout and conventions optimized for LLMs:
   - Flat, explicit structure
   - Minimal “magic”
   - Colocated migrations/schema + handlers
   - Clear naming
3) Implement the scaffold in code in the workspace:
   - Root folders (apps/, shared/, deploy/, scripts/, templates/, docs/)
   - Shared packages/modules for:
     - Authentication middleware (cookie/JWT or sessions)
     - Database helper (SQLite open, migrations runner)
     - Base HTML templates/layout (Tailwind + HTMX)
     - Shared models/types
   - A reusable **app template** under templates/ that new apps are generated from.
   - A **new app generator script** (e.g., scripts/new-app.sh) that:
     - Creates apps/{app-name}
     - Applies placeholders {APP_NAME}, {APP_PORT}
     - Initializes go.mod / composer.json / requirements.txt as needed
     - Adds the app to workspace tooling (e.g., go.work)
     - Prints next steps
4) Local development workflow:
   - Makefile targets (dev-<app>, build, up/down, logs)
   - Optional hot reload tooling (keep minimal)
5) Deployment:
   - Docker Compose setup for running multiple apps on one server
   - Reverse proxy config (nginx or caddy) with path-based routing (e.g., /shopping, /expenses)
   - Environment variables via deploy/.env.example
   - Backup script for per-app databases (esp. SQLite)
6) Provide a fully working example app that demonstrates the patterns:
   - Authenticated pages
   - CRUD with HTMX
   - SQLite migrations
   - Minimal but clean UI
7) Verification:
   - Commands to build/run locally
   - Commands to build docker images
   - Minimal “smoke test” checklist

Constraints:
- Keep it **simple** and **low-maintenance**.
- Avoid unnecessary complexity (no Kubernetes unless explicitly requested).
- Prefer isolation boundaries that prevent cross-app breakage.
- Don’t invent extra UX beyond what’s requested; focus on a solid baseline.

Output format:
- First: short decision summary (stack + deploy + DB + repo strategy).
- Then: step-by-step implementation actions and file changes.
- Then: commands to run locally and deploy.
- Ask at most 1–3 clarifying questions only if truly required.
