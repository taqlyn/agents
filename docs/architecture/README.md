# Taqlyn agents

**Date:** 2026-08-15  
**Architecture brief:** A local-first Go MCP authenticates to a Taqlyn account (dashboard session) with **environment** (`sandbox` | `production` | `both`) and **permission** (`read` | `write`). Skills drive integrate/debug so the agent asks the user as little as possible. Shipping form is the **prebuilt image `ghcr.io/taqlyn/agents`** (stdio `docker run -i` or HTTP). Hosted MCP later reuses the same HTTP mode.

## Document map

| Doc | Contents |
|-----|----------|
| [01-context.md](./01-context.md) | Users, clients, Taqlyn API |
| [03-modules.md](./03-modules.md) | MCP modules and tools |
| [05-sequences.md](./05-sequences.md) | Login and integrate flows |
| [08-folder-map.md](./08-folder-map.md) | Paths in this repo |

## Extraction

Keep MCP as one binary / one image. If a hosted multi-tenant MCP ships later, reuse `internal/server` + streamable HTTP and add request-scoped Bearer tokens; do not split tools first.
