# Documentation

This folder contains repo documentation, organized by intent.

## Start here

- **Deployment (droplet)**: [guides/deployment.md](./guides/deployment.md)
- **Ops checklist**: [guides/ops-checklist.md](./guides/ops-checklist.md)
- **Publishing images (build + push)**: [guides/publishing.md](./guides/publishing.md)

## Guides (how-to)

- [guides/deployment.md](./guides/deployment.md) — pull-only droplet deploy with Docker Compose + Caddy
- [guides/publishing.md](./guides/publishing.md) — build + push images (laptop/CI)
- [guides/registry-auth.md](./guides/registry-auth.md) — registry login on a pull-only droplet
- [guides/ops-checklist.md](./guides/ops-checklist.md) — first-time deploy checklist

## Reference (contracts, patterns)

- [reference/design-spec.md](./reference/design-spec.md) — **UI contract** (allowed HTML structure + allowed CSS classes)
- [reference/patterns.md](./reference/patterns.md) — common Go/HTMX/SQLite patterns used in this monorepo
- [reference/llm-prompts.md](./reference/llm-prompts.md) — prompt templates for working with LLM coding assistants

## Architecture notes

- [architecture/deployment-rework-proposal.md](./architecture/deployment-rework-proposal.md) — rationale for the current deploy model (implemented)
- [architecture/architecture-refactor.md](./architecture/architecture-refactor.md) — base-path routing + local Docker notes and follow-up ideas

## App plans / specs

- [plans/projects-plan.md](./plans/projects-plan.md) — design plan for the `projects` app

## UI artifacts

- [ui/ui-kit.html](./ui/ui-kit.html)
- [ui/ui-kit-b.html](./ui/ui-kit-b.html)
- Mockups:
  - [ui/mockups/mockup-a/](./ui/mockups/mockup-a/)
  - [ui/mockups/mockup-b/](./ui/mockups/mockup-b/)

## Blog / long-form notes

These are historical writeups and may describe earlier iterations.

- [blog/blog-zero-to-deployed.md](./blog/blog-zero-to-deployed.md)
- [blog/blog-htmx-go-stack.md](./blog/blog-htmx-go-stack.md)
- [blog/blog-llm-assisted-development.md](./blog/blog-llm-assisted-development.md)
- [blog/blog-agent-ui-contract-migration.md](./blog/blog-agent-ui-contract-migration.md)
