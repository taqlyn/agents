# Taqlyn agents

Go **MCP** plus **agent skills** so an assistant can log into a Taqlyn account, bind apps, fetch sandbox public credentials, and wire SDKs with almost no extra prompting.

Remote/server MCP is the same HTTP mode; this repo ships **local stdio** and **Docker HTTP**.

## What you get

- Account login (`auth_login`) stored at `~/.config/taqlyn/mcp.json` (mode 0600). Password is never written.
- Scopes: **environment** `sandbox` | `production` | `both`, **permission** `read` | `write`. Production write requires `confirmProductionWrite`.
- Tools for apps, platform binds, public credentials, links, and click/match stats.
- Local `inspect_workspace` / `integration_plan` so the agent can detect Android/iOS/RN/Flutter before asking questions.
- Skills: `taqlyn-mcp`, `taqlyn-integrate`, `taqlyn-debug`.

Architecture: [docs/architecture/README.md](docs/architecture/README.md).

## Run locally (stdio)

Needs Go 1.25+ and a Taqlyn API (`http://127.0.0.1:8080` from platform compose, or a hosted API).

```bash
cd packages/agents   # or clone github.com/taqlyn/agents
go run ./cmd/taqlyn-mcp
```

Cursor MCP config:

```json
{
  "mcpServers": {
    "taqlyn": {
      "command": "go",
      "args": ["run", "./cmd/taqlyn-mcp"],
      "cwd": "/ABS/PATH/to/agents",
      "env": {
        "TAQLYN_API_URL": "http://127.0.0.1:8080",
        "TAQLYN_MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

Then in chat: the agent should call `auth_login` (sandbox + write is enough to integrate).

## Docker (HTTP)

```bash
cp .env.example .env   # set TAQLYN_API_URL if needed
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

`GET /healthz` is the process probe. Token file lives in the `taqlyn_mcp_config` volume.

## Skills

```bash
./scripts/install-skills.sh
```

Or copy `skills/` into the consuming app as `.cursor/skills/`.

## Tests

```bash
go test ./...
```
