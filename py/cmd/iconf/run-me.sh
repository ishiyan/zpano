#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../.."
PYTHONPATH=. python3 -m py.cmd.iconf py/cmd/iconf/settings.json /tmp/iconf_py_output
