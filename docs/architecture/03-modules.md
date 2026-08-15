# Modules

| Module | Interface | Owns |
|--------|-----------|------|
| `internal/auth` | Snapshot + env/permission guards | `~/.config/taqlyn/mcp.json` |
| `internal/api` | HTTP wrapper | No persistence |
| `internal/workspace` | `Inspect(root)` | Nothing remote |
| `internal/tools` | MCP tool handlers | None — call api/auth/workspace |
| `internal/server` | stdio or streamable HTTP | Process |

Tools never import `net/http`. Feature/skill markdown never names vendor MCP types.

**Scope matrix**

| Tool class | read + sandbox | write + sandbox | production |
|------------|----------------|-----------------|------------|
| inspect / health | yes (no login for inspect) | yes | n/a |
| list/get apps, links, stats | same env only | same | needs env scope |
| create/bind/create_link | denied | yes | needs env + write; prod write extra confirm |
