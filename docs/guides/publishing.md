# Publishing images (build + push)

This project is designed so the droplet is **pull-only**:
- build + push images on your laptop or in CI
- deploy on the droplet by pulling images and restarting containers

## 1) Prerequisites

- Docker installed locally
- Logged into your registry (if private)

Example (DigitalOcean Container Registry):

```bash
docker login registry.digitalocean.com
```

## 2) Build + push

Use `scripts/publish-images.sh`:

```bash
export REGISTRY=registry.digitalocean.com/<your-registry>
export IMAGE_TAG=latest

./scripts/publish-images.sh
```

## 3) Deploy on the droplet

After publishing, SSH to the droplet and run:

```bash
cd /opt/a-apps
./scripts/deploy.sh
```

(or run the compose commands directly from `/opt/a-apps/deploy` – see [deployment.md](./deployment.md)).
