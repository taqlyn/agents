# Taqlyn agents

**Date:** 2026-08-15  
**Architecture brief:** A Go MCP (prebuilt `ghcr.io/taqlyn/agents`) authenticates to the **live** Taqlyn HTTPS API with environment (`sandbox` | `production` | `both`) and permission (`read` | `write`). Local APIs are rejected. Install: `curl -fsSL https://raw.githubusercontent.com/taqlyn/agents/main/scripts/install.sh | bash`.

## Document map

| Doc | Contents |
|-----|----------|
| [01-context.md](./01-context.md) | Users, clients, Taqlyn API |
| [03-modules.md](./03-modules.md) | MCP modules and tools |
| [05-sequences.md](./05-sequences.md) | Login and integrate flows |
| [08-folder-map.md](./08-folder-map.md) | Paths in this repo |

## Extraction

Keep MCP as one binary / one image. If a hosted multi-tenant MCP ships later, reuse `internal/server` + streamable HTTP and add request-scoped Bearer tokens; do not split tools first.
