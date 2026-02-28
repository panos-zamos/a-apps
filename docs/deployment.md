# Deployment (Droplet) – pull-only + Docker Compose + Caddy

The droplet is **pull-only**:
- images are built + pushed on your laptop/CI
- the droplet only **pulls** images and restarts containers

This keeps the server simple and avoids surprise builds.

## 0) Prerequisites

1. A Linux server (e.g. DigitalOcean droplet)
2. DNS A records pointing to the droplet IP:
   - `holisticagency.com` → `<droplet-ip>`
3. Firewall allows inbound:
   - TCP 80 (HTTP) – required for ACME HTTP-01 validation
   - TCP 443 (HTTPS)
4. Docker + Docker Compose plugin installed on the droplet

## 1) First-time droplet setup

### 1.1 Clone the repo

```bash
mkdir -p /opt/a-apps
cd /opt/a-apps
# git clone <repo> .
```

### 1.2 Create `deploy/.env`

```bash
cd /opt/a-apps
cp deploy/.env.example deploy/.env
nano deploy/.env
```

Required variables (no defaults):
- `DOMAIN` – the public hostname for Caddy + HTTPS cert issuance
- `REGISTRY` – registry prefix to pull from
- `IMAGE_TAG` – tag to pull
- `JWT_SECRET` – shared auth secret (must be stable)

### 1.3 Registry login (only if registry/repo is private)

If your registry/repo is private, authenticate once on the droplet. Docker stores credentials in `~/.docker/config.json` and `docker compose pull` will reuse them.

Example (DigitalOcean):

```bash
docker login registry.digitalocean.com
# or: doctl registry login
```

## 2) Deploy on the droplet

### Option A: the simplest (recommended)

```bash
cd /opt/a-apps
./scripts/deploy.sh
```

This runs `scripts/deploy-local.sh` which:
- validates required variables from `deploy/.env`
- runs compose from `deploy/` (implicit `.env`)

### Option B: run compose directly

```bash
cd /opt/a-apps/deploy

docker compose pull
docker compose up -d --remove-orphans
docker compose ps
```

## 3) Verify HTTPS / certificate issuance

```bash
docker logs -f a-apps-caddy
```

If ACME fails, double check:
- DNS points to the droplet
- ports 80/443 reachable from the internet

(See also: `docs/ops-checklist.md` for a first-time deploy checklist.)

## 4) Troubleshooting TLS (ERR_SSL_PROTOCOL_ERROR)

From anywhere:

```bash
openssl s_client -connect holisticagency.com:443 -servername holisticagency.com
```

If you see `no peer certificate available`, the server is not presenting a certificate for that hostname.
Most often this means:
- `DOMAIN` is missing/wrong in `deploy/.env`, or
- ports 80/443 are blocked so ACME cannot validate.

Fix:
1) ensure `deploy/.env` contains `DOMAIN=holisticagency.com`
2) restart Caddy:

```bash
cd /opt/a-apps/deploy
docker compose up -d --force-recreate caddy
```
