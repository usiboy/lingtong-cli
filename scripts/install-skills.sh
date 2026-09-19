#!/usr/bin/env bash
# Copyright (c) 2026 Lingtong
# SPDX-License-Identifier: MIT
#
# Thin wrapper around `lingtong-cli skills install`.
#
# It deploys the embedded Agent Skills into every AI coding editor it can find
# (Claude Code, OpenCode, Qoder, Cursor, Trae, Codex) so those agents discover
# the CLI, its commands, and its skills. All arguments are forwarded verbatim to
# `lingtong-cli skills install`, e.g.:
#
#   ./scripts/install-skills.sh
#   ./scripts/install-skills.sh --editor claude,opencode
#   ./scripts/install-skills.sh --scope global
#   ./scripts/install-skills.sh --dry-run
#
# Resolution order for the binary:
#   1. $LINGTONG_CLI if set
#   2. lingtong-cli on PATH
#   3. ./lingtong-cli built from source (runs `make build` if missing)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

resolve_cli() {
  if [[ -n "${LINGTONG_CLI:-}" && -x "${LINGTONG_CLI}" ]]; then
    echo "${LINGTONG_CLI}"; return 0
  fi
  if command -v lingtong-cli >/dev/null 2>&1; then
    command -v lingtong-cli; return 0
  fi
  if [[ -x "${SCRIPT_DIR}/lingtong-cli" ]]; then
    echo "${SCRIPT_DIR}/lingtong-cli"; return 0
  fi
  # Build from source as a last resort.
  if command -v make >/dev/null 2>&1 && [[ -f "${SCRIPT_DIR}/Makefile" ]]; then
    echo "lingtong-cli not found; building from source..." >&2
    make -C "${SCRIPT_DIR}" build >&2
    echo "${SCRIPT_DIR}/lingtong-cli"; return 0
  fi
  echo "error: lingtong-cli not found. Install it (npm i -g @lingtong-cli/cli, or 'make install') and retry." >&2
  return 1
}

CLI="$(resolve_cli)"
exec "${CLI}" skills install "$@"
