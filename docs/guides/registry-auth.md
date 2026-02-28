# Registry authentication (pull-only droplet)

On the droplet this project is intended to be **pull-only**:
- images are built and pushed elsewhere (laptop or CI)
- the droplet only runs `docker compose pull` and `docker compose up -d`

## Do we need registry login to pull?

It depends:

- **Public registry + public repo**: no login required.
- **Private registry or private repo**: yes, login is required. Otherwise `docker pull` / `docker compose pull` fails with `401 Unauthorized` / `denied`.

Docker Compose itself does not implement auth; it relies on Docker’s normal credential mechanism.

## What exactly is needed

One of the following credential methods must be present on the droplet:

1) **A successful `docker login <registry-host>`** for the user that runs Docker.
   - Stores credentials in `~/.docker/config.json`.
   - Works with passwords or tokens (recommended: token).

2) A configured **docker credential helper** (less common on servers).

3) Registry-specific tooling that performs the login (e.g. **`doctl registry login`** for DigitalOcean).

## Simplest setup (recommended)

Login once on the droplet and let Docker persist credentials:

```bash
# As root
# (or as a user in the docker group, if that is how your droplet is set up)
docker login registry.digitalocean.com
```

Then deploy without ever re-supplying credentials:

```bash
cd /opt/a-apps/deploy
docker compose pull
docker compose up -d --remove-orphans
```

## Security notes

- Prefer a **read-only** token if your registry supports it.
- Avoid echoing tokens in scripts/logs.
- If you do store tokens in `deploy/.env`, ensure:
  - the file is not committed (it is ignored via `.gitignore`)
  - permissions are restrictive (e.g. `chmod 600 deploy/.env`)

## Quick troubleshooting

Check whether Docker has credentials:

```bash
cat ~/.docker/config.json
```

Test a pull:

```bash
docker pull registry.digitalocean.com/<your-registry>/a-apps-todo-list:latest
```
