# A-Apps: Personal Web Applications Platform

A monorepo for hosting multiple small web applications with minimal friction. Built with Go + HTMX for simplicity and rapid development with LLM assistance.

## Architecture

- **Stack:** Go + HTMX + SQLite + Docker Compose
- **Isolation:** Each app is independent with its own database and container
- **Shared:** Authentication, UI components, deployment configuration
- **Deployment:** Single Digital Ocean Droplet with Caddy reverse proxy

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Make (optional, for convenience commands)

### Development

1. **Create a new app:**
   ```bash
   ./scripts/new-app.sh my-app 3005
   ```

2. **Run locally:**
   ```bash
   cd apps/my-app
   go run main.go
   # Or with hot reload
   make dev-my-app
   ```

3. **Access:** http://localhost:3005

### Run All Apps with Docker Compose (Production-like)

`deploy/docker-compose.yml` is intended for a production-style setup where the droplet **pulls pre-built images from a registry**.

From the repo root:

```bash
cp deploy/.env.example deploy/.env
# edit deploy/.env (DOMAIN, REGISTRY, IMAGE_TAG, JWT_SECRET, ...)

docker compose --env-file deploy/.env -f deploy/docker-compose.yml pull
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --remove-orphans
```

Then access your apps at:

- http://localhost/todo/
- http://localhost/projects/

The compose file sets `BASE_PATH` for each app so redirects and assets work behind Caddy.

Stop everything with:

```bash
docker compose --env-file deploy/.env -f deploy/docker-compose.yml down
```

Caddy will automatically provision HTTPS certificates for real domains as long as `DOMAIN` in `deploy/.env` is set to the public hostname and ports 80/443 are reachable.

### Deployment

See **docs/deployment.md** for the step-by-step droplet setup.

At a minimum you must create `deploy/.env` on the droplet (copy from `deploy/.env.example`) and set:
- `DOMAIN`
- `REGISTRY`
- `IMAGE_TAG`
- `JWT_SECRET`

Then:
- Run `./scripts/deploy.sh` **on the droplet** to pull + restart containers, or
- Run `./scripts/deploy.sh` **from your laptop** (remote mode) to build+push images, sync code, then restart containers on the droplet.

## Directory Structure

```
a-apps/
├── apps/                    # Individual applications
│   ├── todo-list/          # Example: Multi-store shopping list
│   ├── expense-tracker/    # Example: Personal expense tracking
│   └── project-tracker/    # Example: Hobby project progress
├── shared/                  # Shared code across apps
│   ├── auth/               # Authentication middleware
│   ├── database/           # SQLite helpers
│   ├── templates/          # Base HTML layouts
│   └── models/             # Common types
├── deploy/                  # Deployment configuration
│   ├── docker-compose.yml  # Production compose file
│   ├── Caddyfile           # Reverse proxy config
│   └── .env.example        # Environment variables template
├── scripts/                 # Automation scripts
│   ├── new-app.sh          # Scaffold new application
│   ├── deploy.sh           # Deploy to server
│   ├── backup.sh           # Backup SQLite databases
│   └── hash-password.sh    # Generate password hashes
├── docs/                    # Documentation
│   ├── llm-prompts.md      # Effective prompts for LLM assistance
│   └── patterns.md         # Common Go/HTMX patterns
├── templates/               # App template for scaffolding
│   └── app-template/       # Base structure for new apps
└── go.work                  # Go workspace configuration
```

## Adding a New App

1. Run the scaffold script:
   ```bash
   ./scripts/new-app.sh expense-tracker 3002
   ```

2. The script will:
   - Create `apps/expense-tracker/` from template
   - Initialize SQLite database
   - Add service to `deploy/docker-compose.yml`
   - Update `deploy/Caddyfile`
   - Assign port 3002

3. Customize your app:
   - Modify `handlers/` for your business logic
   - Update `templates/` for your UI
   - Add tables to `db/migrations/`
   - Configure users in `config.yaml`

4. Test locally:
   ```bash
   cd apps/expense-tracker
   go run main.go
   ```

## Authentication

Each app uses shared authentication middleware with config-based users:

```yaml
# apps/my-app/config.yaml
users:
  - username: panos
    password_hash: $2a$10$...  # bcrypt hash
  - username: guest
    password_hash: $2a$10$...
```

Generate password hashes:
```bash
./scripts/hash-password.sh mypassword
```

## LLM-Assisted Development

This project structure is optimized for working with LLM coding assistants:

- **Flat structure:** Easy for LLMs to understand full context
- **Explicit code:** Minimal magic, clear dependencies
- **Reusable patterns:** Reference existing apps as examples
- **Template-based:** Generate new apps using established patterns

See [docs/llm-prompts.md](docs/llm-prompts.md) for effective prompts.

## Technology Choices

- **Go:** Single binary deployment, fast compilation, excellent for CRUD apps
- **HTMX:** Dynamic UIs without heavy JavaScript frameworks
- **SQLite:** Perfect for single-user apps, zero configuration, file-based backups
- **Docker Compose:** Simple orchestration, easy deployment
- **Caddy:** Reverse proxy with automatic HTTPS

## Deployment Architecture

```
┌─────────────────────────────────────┐
│     Caddy (Port 80/443)             │
│  ┌─────────────────────────────┐   │
│  │ /todo → localhost:3001       │   │
│  │ /projects → localhost:3002   │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
         │           │
    ┌────▼───┐  ┌───▼────┐
    │ App 1  │  │ App 2  │
    │ :3001  │  │ :3002  │
    └────┬───┘  └───┬────┘
         │          │
    ┌────▼───┐  ┌───▼────┐
    │ SQLite │  │ SQLite │
    └────────┘  └────────┘
```

## Future Enhancements

- [ ] Shared authentication service (single sign-on)
- [ ] Automated HTTPS with Let's Encrypt
- [ ] GitHub Actions CI/CD
- [ ] Monitoring dashboard
- [ ] Database migration tooling
- [ ] App marketplace/catalog UI
- [ ] Shared component library
- [ ] API gateway for mobile apps

## License

MIT - Personal use project
