# Deployment rework proposal (simpler + idiomatic)

Status: **implemented**.

This document explains the rationale behind the simplified deployment workflow (pull-only droplet, explicit `deploy/.env`, separate publish vs deploy).

## Goals

- **One obvious way** to deploy.
- **No hidden defaults** (fail fast if something is missing).
- **Single source of truth** for deploy configuration.
- Keep scripts **small**, **composable**, and easy to audit.

## Current deployment (what it does today)

### Runtime topology

- A single droplet runs `deploy/docker-compose.yml`.
- `caddy` listens on **80/443** and reverse-proxies to apps.
- Caddy config uses the `DOMAIN` variable in `deploy/Caddyfile`:

  - If `DOMAIN` is incorrect/missing, Caddy may not request a cert for the real hostname.
  - Symptom can be TLS handshake failures like `ERR_SSL_PROTOCOL_ERROR` / `no peer certificate available`.

### Configuration

- `deploy/.env` is used for docker compose variable substitution.
- Variables in practice:
  - `DOMAIN` (Caddy site + ACME cert)
  - `REGISTRY` / `IMAGE_TAG` (which images to pull)
  - `JWT_SECRET` (shared auth secret)
  - optional registry credentials (if private registry)

### Automation

- There is a single `scripts/deploy.sh` that tries to support:
  - local deploy on the droplet
  - remote deploy from a laptop (build + push + sync + restart)

This provides convenience but increases complexity because it needs branching logic, defaults, inference, and cross-machine coordination.

## Key simplification idea: "deploy happens on the droplet"

Make the droplet the place where deployment is executed and validated.

- The droplet owns **`deploy/.env`**.
- The droplet runs **`docker compose pull && docker compose up -d`**.
- Your laptop (or CI) only **publishes images**.

This separation is idiomatic and reduces surprises:
- publishing images is one concern
- pulling and running containers is another

## Proposed (simplified) operational model

### 1) Canonical working directory: `deploy/`

Docker Compose automatically reads `.env` from the **current working directory**.

**Rule:** always run compose commands from `deploy/`.

Example on the droplet:

```bash
cd /opt/a-apps/deploy

# uses ./deploy/.env implicitly
docker compose pull
docker compose up -d --remove-orphans
docker compose ps
```

Why this is simpler:
- no `--env-file` flags to remember
- no ambiguity about which `.env` is used
- fewer moving parts

### 2) No defaults; explicit requirements

In docs and scripts (if you later choose to refactor), treat these as mandatory in production:

- `DOMAIN`
- `REGISTRY`
- `IMAGE_TAG`
- `JWT_SECRET`

Rationale:
- `DOMAIN` determines whether HTTPS works at all.
- `REGISTRY`/`IMAGE_TAG` determine which build is running.
- `JWT_SECRET` affects auth validity and must be stable.

### 3) Replace one big script with two tiny scripts (composable)

If you decide to refactor later, split responsibilities:

#### A. Publish script (runs on laptop/CI)
Responsibility: **build + push images**.

Interface:
- requires: `REGISTRY`, `IMAGE_TAG`
- optional: `REGISTRY_USER`, `REGISTRY_TOKEN` (or rely on existing docker login)

Pseudo-usage:
```bash
REGISTRY=... IMAGE_TAG=... ./scripts/publish-images.sh
```

#### B. Deploy script (runs on droplet)
Responsibility: **pull + restart containers**.

Interface:
- requires: `deploy/.env` exists
- runs from `deploy/` directory

Pseudo-usage:
```bash
cd /opt/a-apps
./scripts/deploy-local.sh   # or: cd deploy && docker compose ...
```

Why this reduces cognitive load:
- each script is ~10–30 lines
- no mode switching, no SSH branching
- fewer variables to thread through

### 4) Prefer a human-readable "deployment contract"

Treat `deploy/.env` as the contract.

Guidelines:
- keep it small
- comment every variable with purpose
- avoid derived values in scripts (e.g. do not infer registry host; specify it if needed)

### 5) Optional: systemd for the last mile (no scripts at all)

Long term, the simplest deploy is:

- publish images (CI)
- ssh to droplet and run `docker compose pull && docker compose up -d`

If you want repeatable boot behavior and log location, you can later wrap compose in a systemd service.

## Suggested documentation layout

To make the workflow obvious, keep docs focused:

- `docs/guides/deployment.md`
  - exact commands to run on the droplet
  - first-time setup
  - troubleshooting

- `deploy/.env.example`
  - the deployment contract with clear explanations

- (optional) `docs/guides/publishing.md`
  - how to build/push images (from laptop or CI)

## Why this is better (summary)

- **Less magic**: fewer inferred/assumed values.
- **Fewer modes**: deploy is always executed on the droplet.
- **Easier auditing**: small scripts or direct compose commands.
- **Fewer footguns**: `DOMAIN` is always explicit; `.env` lookup is deterministic.

## Migration sketch (if you choose to implement later)

1. Keep `deploy/.env` as the sole config.
2. Document the droplet commands as the primary deployment method.
3. Move build+push into a separate publish step (script or CI).
4. Optionally delete complex “remote deploy” logic once the new workflow is trusted.
