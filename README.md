# Taqlyn agents

**MCP talks to live Taqlyn only** (`https://api.rutvik.qzz.io`). It will not use a local API.

## Install (no clone, no Go)

Needs Docker.

```bash
curl -fsSL https://raw.githubusercontent.com/taqlyn/agents/main/scripts/install.sh | bash
```

That pulls `ghcr.io/taqlyn/agents:latest`, writes `~/.cursor/mcp.json`, and installs the Cursor skills. Restart Cursor, then `auth_login`.

Override the live origin if you self-host a public HTTPS API:

```bash
curl -fsSL https://raw.githubusercontent.com/taqlyn/agents/main/scripts/install.sh | TAQLYN_API_URL=https://api.example.com bash
```

Image: `ghcr.io/taqlyn/agents:latest` (linux/amd64, linux/arm64).

If `docker pull` is 401, sign in once (`docker login ghcr.io`) or make the [GHCR package](https://github.com/taqlyn/agents/pkgs/container/agents) public.

## What you get

- Account login (`auth_login`) against the **live** control plane. Session in the `taqlyn-mcp-config` volume (0600). Password is never written.
- Scopes: environment `sandbox` | `production` | `both`; permission `read` | `write`. Production write requires `confirmProductionWrite`.
- Tools for apps, binds, public credentials, links, and click/match stats.
- `inspect_workspace` / `integration_plan` read the **project files** mounted at `/workspace`. Account data always comes from the live API.

Architecture: [docs/architecture/README.md](docs/architecture/README.md).

## HTTP (same live API)

```bash
docker run --rm -p 8787:8787 \
  -e TAQLYN_MCP_TRANSPORT=http \
  -e TAQLYN_API_URL=https://api.rutvik.qzz.io \
  -v taqlyn-mcp-config:/home/mcp/.config/taqlyn \
  ghcr.io/taqlyn/agents:latest
```

```json
{ "mcpServers": { "taqlyn": { "url": "http://127.0.0.1:8787/mcp" } } }
```

## Contributors

```bash
go test ./...
docker compose -f docker-compose.yml -f docker-compose.build.yml up --build
```
