#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
dest="${1:-$HOME/.cursor/skills}"
mkdir -p "$dest"
cp -R "$root/skills/." "$dest/"
echo "Installed Taqlyn skills into $dest"
