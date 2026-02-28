# Deployment (Droplet) – Caddy + Docker Compose

This project is deployed as Docker containers behind **Caddy** (HTTPS) using `deploy/docker-compose.yml`.

The most common TLS failure on a new setup is running Caddy with the default `DOMAIN=localhost`. In that case Caddy will not request a certificate for your real hostname and clients can see errors like `ERR_SSL_PROTOCOL_ERROR`.

## 0) Prerequisites

1. A Linux server (e.g. DigitalOcean droplet)
2. DNS A records pointing to the droplet IP:
   - `holisticagency.com` → `<droplet-ip>`
   - (optional) `www.holisticagency.com` → `<droplet-ip>`
3. Firewall allows inbound:
   - TCP 80 (HTTP) – required for ACME HTTP-01 validation
   - TCP 443 (HTTPS)
4. Docker + Docker Compose plugin installed on the droplet

## 1) First-time setup on the droplet

### 1.1 Clone the repo

```bash
mkdir -p /opt/a-apps
cd /opt/a-apps
# clone / rsync your repo here
```

### 1.2 Create `deploy/.env`

Copy the example and edit values:

```bash
cd /opt/a-apps
cp deploy/.env.example deploy/.env
nano deploy/.env
```

Minimum required values:

- `DOMAIN`
  - Must be the **public hostname** you will use in the browser (e.g. `holisticagency.com`).
  - Caddy uses this to decide which site to serve and which ACME certificate to request.

- `REGISTRY`
  - Registry prefix where images are pushed (e.g. `registry.digitalocean.com/<your-registry>`).

- `IMAGE_TAG`
  - Tag to deploy (usually `latest` or a CI build tag).

- `JWT_SECRET`
  - Must be the same for all app containers.
  - Generate a strong one:
    ```bash
    openssl rand -hex 32
    ```

If the registry is private, also set:

- `REGISTRY_USER`
- `REGISTRY_TOKEN`
- `REGISTRY_HOST` (optional; can be derived from `REGISTRY`)

### 1.3 (Optional) Authenticate the droplet to your registry

If you don’t want the deploy script to do registry login, you can do it once manually:

```bash
# example for DO registry host
docker login registry.digitalocean.com
```

## 2) Deploy manually on the droplet (recommended for debugging)

From the repo root:

```bash
cd /opt/a-apps

# Pull images defined by REGISTRY/IMAGE_TAG
docker compose --env-file deploy/.env -f deploy/docker-compose.yml pull

# Start/upgrade containers
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --remove-orphans

# Check status
docker compose --env-file deploy/.env -f deploy/docker-compose.yml ps
```

### 2.1 Verify HTTPS / certificate issuance

Caddy logs:

```bash
docker logs -f a-apps-caddy
```

If ACME fails, double check:
- DNS points to the droplet
- ports 80/443 reachable from the internet

## 3) Using `scripts/deploy.sh`

`scripts/deploy.sh` supports:
- running **on the droplet** (no SSH required)
- running **from your laptop** to deploy to a remote droplet over SSH

The script expects configuration in environment variables and/or `deploy/.env`.

### 3.1 Run on the droplet

```bash
cd /opt/a-apps
./scripts/deploy.sh
```

### 3.2 Run from your laptop (remote deploy)

High-level flow:
1. build + push images to `REGISTRY`
2. rsync code to the droplet (without secrets)
3. ssh into droplet and run docker compose using the droplet’s `deploy/.env`

Example:

```bash
export DEPLOY_HOST=<droplet-ip-or-host>
export DEPLOY_USER=root
export DEPLOY_PATH=/opt/a-apps

# Needed to push images from your laptop
export REGISTRY_USER=you@example.com
export REGISTRY_TOKEN=...

# REGISTRY and IMAGE_TAG are read from the droplet's deploy/.env to avoid mismatches.
# (You may still export them locally, but they must match the server values.)

./scripts/deploy.sh
```

## 4) Troubleshooting TLS (ERR_SSL_PROTOCOL_ERROR)

Run from anywhere:

```bash
openssl s_client -connect holisticagency.com:443 -servername holisticagency.com
```

If you see `no peer certificate available`, the server is not presenting a certificate for that hostname.
Most often this means `DOMAIN` is not set correctly (Caddy configured for `localhost`).

Fix: ensure `deploy/.env` contains `DOMAIN=holisticagency.com`, then redeploy/restart Caddy:

```bash
cd /opt/a-apps
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up -d --force-recreate caddy
```
