#!/usr/bin/env bash
set -euo pipefail

read -r -p "Go live? [y/N]: " go_live
if [[ "${go_live}" =~ ^[Yy]$ ]]; then
  export TRADING_MODE=live
else
  export TRADING_MODE=paper
fi

mkdir -p data/logs
log_file="data/logs/runtime_$(date -u +%Y%m%dT%H%M%SZ).log"

go_bin="${GO_BIN:-go}"
if ! command -v "${go_bin}" >/dev/null 2>&1; then
  if [[ -x "/c/Program Files/Go/bin/go.exe" ]]; then
    go_bin="/c/Program Files/Go/bin/go.exe"
  else
    echo "go executable not found in this shell. Set GO_BIN or run scripts/run_live_logged.ps1 from PowerShell." >&2
    exit 127
  fi
fi

"${go_bin}" run ./cmd/main.go 2>&1 | tee "${log_file}"
