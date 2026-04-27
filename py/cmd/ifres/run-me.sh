#!/usr/bin/env bash
#
# Run with: PYTHONPATH=. python -m py.cmd.ifres py/cmd/ifres/settings.json from the project root.

PYTHONPATH=. python3 -m py.cmd.ifres py/cmd/ifres/settings.json >output.txt
