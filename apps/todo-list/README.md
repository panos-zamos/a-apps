# todo-list

A personal web application built with Go + HTMX.

## Development

```bash
# Run locally
go run main.go

# Access at http://localhost:3001
```

## Configuration

Edit `config.yaml` to add users:

```yaml
users:
  - username: yourname
    password_hash: $2a$10$...  # Generate with ../../scripts/hash-password.sh
```

## Building

```bash
# Build binary
go build -o todo-list .

# Run
./todo-list
```

## Docker

```bash
# Build image
docker build -t todo-list .

# Run container
docker run -p 3001:3001 -v $(pwd)/data:/root/data todo-list
```
