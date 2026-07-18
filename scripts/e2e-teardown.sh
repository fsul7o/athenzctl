#!/usr/bin/env bash
# Kill port-forward processes started by e2e-bootstrap.sh.
# Usage: e2e-teardown.sh <out-dir>
set -euo pipefail

OUT_DIR="${1:?output dir required}"
PID_FILE="${OUT_DIR}/pids"
if [ -f "${PID_FILE}" ]; then
  while read -r pid; do
    [ -n "${pid}" ] || continue
    kill "${pid}" 2>/dev/null || true
  done <"${PID_FILE}"
  rm -f "${PID_FILE}"
fi
