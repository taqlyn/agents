---
name: taqlyn-mcp
description: Installs and authenticates the Taqlyn MCP against live Taqlyn (curl | bash, Docker image). Use when setting up Taqlyn MCP, Cursor mcp.json, install.sh, login, or sandbox/production and read/write scopes. Never point it at a local API.
---

# Taqlyn MCP

**Live API only** (`https://api.rutvik.qzz.io`). Do not configure localhost or `host.docker.internal`.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/taqlyn/agents/main/scripts/install.sh | bash
```

Restart Cursor, then `auth_login`. The installer writes `~/.cursor/mcp.json` and skills under `~/.cursor/skills/`.

Self-hosted HTTPS origin:

```bash
curl -fsSL https://raw.githubusercontent.com/taqlyn/agents/main/scripts/install.sh | TAQLYN_API_URL=https://api.example.com bash
```

The MCP process is `ghcr.io/taqlyn/agents:latest` over stdio. `inspect_workspace` uses `/workspace` (the Cursor folder). Account operations always hit the live API.

If `docker pull` is 401, `docker login ghcr.io` once.

## Scopes

| Dimension | Values | Notes |
|-----------|--------|--------|
| environment | `sandbox`, `production`, `both` | Default sandbox |
| permission | `read`, `write` | Default write for integrate |
| production write | `confirmProductionWrite: true` | Required |

Then [taqlyn-integrate](../taqlyn-integrate/SKILL.md) and [taqlyn-debug](../taqlyn-debug/SKILL.md).
