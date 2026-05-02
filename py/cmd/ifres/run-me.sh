#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../.."
PYTHONPATH=. python3 -m py.cmd.ifres py/cmd/ifres/settings.json >py/cmd/ifres/output.txt
