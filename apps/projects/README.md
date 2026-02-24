# Projects

Project tracker application for planning and tracking projects.

## Features

- Track projects with name, description, URLs, stage, and rating
- Filter by stage, type (commercial/open-source/public), and rating
- Timeline/research log with nested entries
- Two separate URL fields: website and source code

## Running

```bash
cd apps/projects
go run .
```

Default port: 3002

## Configuration

Edit `config.yaml` to manage users. Generate password hashes with:

```bash
./scripts/hash-password.sh yourpassword
```
