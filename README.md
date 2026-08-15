# Taqlyn agents

**MCP is a published Docker image.** You do not need Go, and you do not need to clone this repo.

```text
ghcr.io/taqlyn/agents:latest
```

linux/amd64 and linux/arm64. Cursor (or any MCP host) runs the container over stdio. The same image can listen on HTTP for Docker Compose / a future server.

Architecture: [docs/architecture/README.md](docs/architecture/README.md).

## Cursor — pull and run (recommended)

One time:

```bash
docker pull ghcr.io/taqlyn/agents:latest
```

Anonymous pull works once the GitHub Container package is **public** ([package page](https://github.com/taqlyn/agents/pkgs/container/agents) → Change visibility). If the pull is 401, sign in once:

```bash
echo YOUR_GITHUB_TOKEN | docker login ghcr.io -u YOUR_GITHUB_USER --password-stdin
docker pull ghcr.io/taqlyn/agents:latest
```

`~/.cursor/mcp.json` or the project `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "taqlyn": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
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

Point `TAQLYN_API_URL` at your control plane (`http://host.docker.internal:8080` for platform compose, or `https://api.example.com` for hosted). Then `auth_login` (sandbox + write is enough to integrate).

Inspect/plan tools use `/workspace` automatically when that mount is present.

## HTTP (no stdio)

```bash
docker run --rm -p 8787:8787 \
  --add-host=host.docker.internal:host-gateway \
  -e TAQLYN_MCP_TRANSPORT=http \
  -e TAQLYN_API_URL=http://host.docker.internal:8080 \
  -v taqlyn-mcp-config:/home/mcp/.config/taqlyn \
  ghcr.io/taqlyn/agents:latest
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

Or `docker compose up` in this repo (pulls the same image). `GET /healthz` is the HTTP probe.

## Scopes

| Dimension | Values | Notes |
|-----------|--------|--------|
| environment | `sandbox`, `production`, `both` | Default sandbox |
| permission | `read`, `write` | Default write for integrate |
| production write | `confirmProductionWrite: true` | Required |

The session is stored in the `taqlyn-mcp-config` volume (`mcp.json`, mode 0600). Password is never written. Public `clientId` / `publicKeyId` only — no private PEM.

## Skills

From the image users still copy skills onto the machine that runs Cursor (MCP tools plus these files). If this repo is already a submodule:

```bash
./scripts/install-skills.sh
```

Else fetch the three `SKILL.md` files from [`skills/`](skills/) on GitHub.

## Contributors (Go / local image)

```bash
go test ./...
docker compose -f docker-compose.yml -f docker-compose.build.yml up --build
```
