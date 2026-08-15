---
name: taqlyn-mcp
description: Runs and authenticates the Taqlyn Go MCP (stdio or Docker HTTP) so agents can operate a Taqlyn account with sandbox/production and read/write scopes. Use when setting up Taqlyn MCP, Cursor mcp.json, Docker, or login scopes.
---

# Taqlyn MCP

Go MCP (`taqlyn-mcp`). Local stdio or HTTP; Docker for HTTP. Remote hosting is later — same binary.

## Cursor (stdio)

From this repo:

```bash
go install github.com/taqlyn/agents/cmd/taqlyn-mcp@latest
```

Or build: `go build -o taqlyn-mcp ./cmd/taqlyn-mcp`

`~/.cursor/mcp.json` (or project `.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "taqlyn": {
      "command": "taqlyn-mcp",
      "env": {
        "TAQLYN_API_URL": "http://127.0.0.1:8080",
        "TAQLYN_MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

Hosted API example: `TAQLYN_API_URL=https://api.rutvik.qzz.io`

## Docker (HTTP)

```bash
docker compose up --build
```

```json
{
  "mcpServers": {
    "taqlyn": {
      "url": "http://127.0.0.1:8787/mcp"
    }
  }
}
```

Do not commit passwords. Prefer `auth_login` after connect.

## Scopes

| Dimension | Values | Notes |
|-----------|--------|--------|
| environment | `sandbox`, `production`, `both` | Default sandbox |
| permission | `read`, `write` | Default write for integrate |
| production write | `confirmProductionWrite: true` | Required |

Token file: `$XDG_CONFIG_HOME/taqlyn/mcp.json` (0600). Password is never stored.

## Skills on disk

Copy `skills/*` into the consumer app:

```bash
./scripts/install-skills.sh
# or: cp -R skills/* ~/.cursor/skills/
```

Then use [taqlyn-integrate](../taqlyn-integrate/SKILL.md) and [taqlyn-debug](../taqlyn-debug/SKILL.md).
