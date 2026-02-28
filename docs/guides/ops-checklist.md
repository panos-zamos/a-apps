# Ops checklist (first-time deploy)

A quick, minimal checklist to avoid common deployment issues.

## 1) DNS + networking

- `DOMAIN` in `deploy/.env` matches your public hostname (e.g. `holisticagency.com`).
- DNS A record points to the droplet IP.
- Firewall allows inbound **TCP 80** and **TCP 443**.

Quick checks:
```bash
# from your laptop
nslookup holisticagency.com

# from the droplet
ss -ltnp | grep ':80\|:443'
```

## 2) Docker + Compose

- Docker installed on droplet
- Docker Compose plugin available

```bash
docker version
docker compose version
```

## 3) Registry access (pull-only droplet)

If images are private, login once:
```bash
docker login <registry-host>
# or: doctl registry login
```

Verify a pull:
```bash
docker pull <registry>/<image>:<tag>
```

## 4) Deployment contract

Ensure `deploy/.env` exists and has **no defaults**:
- `DOMAIN`
- `REGISTRY`
- `IMAGE_TAG`
- `JWT_SECRET`

```bash
cd /opt/a-apps
cp deploy/.env.example deploy/.env
nano deploy/.env
```

## 5) Deploy

```bash
./scripts/deploy.sh
```

## 6) TLS verification

```bash
openssl s_client -connect holisticagency.com:443 -servername holisticagency.com
```

If you see `no peer certificate available`, the most common causes are:
- `DOMAIN` missing/wrong in `deploy/.env`
- ports 80/443 blocked
