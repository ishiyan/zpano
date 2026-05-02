#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../.."
PYTHONPATH=. python3 -m py.cmd.icalc py/cmd/icalc/settings.json >py/cmd/icalc/output.txt
