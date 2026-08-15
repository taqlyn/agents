#!/usr/bin/env bash
# Install Taqlyn MCP for Cursor via the prebuilt image (live API only).
#   curl -fsSL https://raw.githubusercontent.com/taqlyn/agents/main/scripts/install.sh | bash
set -euo pipefail

IMAGE="${TAQLYN_MCP_IMAGE:-ghcr.io/taqlyn/agents:latest}"
API_URL="${TAQLYN_API_URL:-https://api.rutvik.qzz.io}"
SKILLS_BASE="${TAQLYN_SKILLS_BASE:-https://raw.githubusercontent.com/taqlyn/agents/main/skills}"
CURSOR_HOME="${CURSOR_HOME:-${HOME}/.cursor}"
MCP_JSON="${CURSOR_HOME}/mcp.json"

say() { printf 'taqlyn-mcp: %s\n' "$*"; }
die() { printf 'taqlyn-mcp: %s\n' "$*" >&2; exit 1; }

command -v docker >/dev/null 2>&1 || die "Docker is required. Install Docker Desktop, then re-run."
docker info >/dev/null 2>&1 || die "Docker is installed but not running."

case "${API_URL}" in
  http://127.*|http://localhost*|https://localhost*|http://host.docker.internal*|https://host.docker.internal*)
    die "TAQLYN_API_URL must be the live HTTPS API (default https://api.rutvik.qzz.io), not a local URL."
    ;;
esac

say "pulling ${IMAGE}"
if ! docker pull "${IMAGE}"; then
  die "docker pull failed. If GHCR returned 401, run: docker login ghcr.io"
fi

mkdir -p "${CURSOR_HOME}/skills"

TMP="$(mktemp)"
trap 'rm -f "${TMP}"' EXIT
cat >"${TMP}" <<EOF
{
  "command": "docker",
  "args": [
    "run",
    "-i",
    "--rm",
    "-e", "TAQLYN_MCP_TRANSPORT=stdio",
    "-e", "TAQLYN_WORKSPACE=/workspace",
    "-e", "TAQLYN_API_URL=${API_URL}",
    "-v", "taqlyn-mcp-config:/home/mcp/.config/taqlyn",
    "-v", "\${workspaceFolder}:/workspace:ro",
    "${IMAGE}"
  ]
}
EOF

if command -v python3 >/dev/null 2>&1; then
  MCP_JSON="${MCP_JSON}" python3 - "${TMP}" <<'PY'
import json, os, pathlib, sys
entry_path = pathlib.Path(sys.argv[1])
entry = json.loads(entry_path.read_text())
dest = pathlib.Path(os.environ["MCP_JSON"])
dest.parent.mkdir(parents=True, exist_ok=True)
data = {}
if dest.exists():
    data = json.loads(dest.read_text() or "{}")
servers = data.get("mcpServers") or {}
servers["taqlyn"] = entry
data["mcpServers"] = servers
dest.write_text(json.dumps(data, indent=2) + "\n")
print(dest)
PY
else
  mkdir -p "$(dirname "${MCP_JSON}")"
  cat >"${MCP_JSON}" <<EOF
{
  "mcpServers": {
    "taqlyn": $(cat "${TMP}")
  }
}
EOF
  say "wrote ${MCP_JSON} (python3 not found; existing MCP servers were not merged)"
fi

for skill in taqlyn-mcp taqlyn-integrate taqlyn-debug; do
  mkdir -p "${CURSOR_HOME}/skills/${skill}"
  curl -fsSL "${SKILLS_BASE}/${skill}/SKILL.md" -o "${CURSOR_HOME}/skills/${skill}/SKILL.md"
done

say "installed Cursor MCP config at ${MCP_JSON}"
say "API ${API_URL} (live only)"
say "restart Cursor, then ask the agent to auth_login"
