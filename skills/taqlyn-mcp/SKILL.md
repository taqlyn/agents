---
name: taqlyn-mcp
description: Runs and authenticates the Taqlyn Go MCP from the prebuilt ghcr.io/taqlyn/agents image (no Go, no clone). Use when setting up Taqlyn MCP, Cursor mcp.json, Docker pull, login, or sandbox/production and read/write scopes.
---

# Taqlyn MCP

**Default: pull `ghcr.io/taqlyn/agents:latest` and run it with Docker.** Do not `go install` or clone this repo unless you are changing the MCP itself.

## Cursor

```bash
docker pull ghcr.io/taqlyn/agents:latest
```

`~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "taqlyn": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--add-host=host.docker.internal:host-gateway",
        "-e", "TAQLYN_MCP_TRANSPORT=stdio",
        "-e", "TAQLYN_WORKSPACE=/workspace",
        "-e", "TAQLYN_API_URL=http://host.docker.internal:8080",
        "-v", "taqlyn-mcp-config:/home/mcp/.config/taqlyn",
        "-v", "${workspaceFolder}:/workspace:ro",
        "ghcr.io/taqlyn/agents:latest"
      ]
    }
  }
}
```

Hosted API: set `TAQLYN_API_URL` to that origin. If `docker pull` returns 401, `docker login ghcr.io` once.

`inspect_workspace` / `integration_plan` with an empty `root` uses `/workspace`.

## HTTP

```bash
docker run --rm -p 8787:8787 -e TAQLYN_MCP_TRANSPORT=http ghcr.io/taqlyn/agents:latest
```

Then `"url": "http://127.0.0.1:8787/mcp"`.

## Scopes

| Dimension | Values | Notes |
|-----------|--------|--------|
| environment | `sandbox`, `production`, `both` | Default sandbox |
| permission | `read`, `write` | Default write for integrate |
| production write | `confirmProductionWrite: true` | Required |

Do not commit passwords. Prefer `auth_login` after connect.

Then use [taqlyn-integrate](../taqlyn-integrate/SKILL.md) and [taqlyn-debug](../taqlyn-debug/SKILL.md).
