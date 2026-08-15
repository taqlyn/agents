# Folder map

```text
cmd/taqlyn-mcp/          # process entry (stdio | http)
internal/
  config/                # env
  auth/                  # scopes + token file
  api/                   # Taqlyn HTTP wrapper
  workspace/             # local project detect
  tools/                 # MCP tool adapters
  server/                # transports
skills/                  # Cursor/Claude skills
  taqlyn-mcp/
  taqlyn-integrate/
  taqlyn-debug/
docs/architecture/
Dockerfile
docker-compose.yml
```
