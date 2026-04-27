#!/usr/bin/env bash
#
# Run with: PYTHONPATH=. python -m py.cmd.icalc py/cmd/icalc/settings.json from the project root.

PYTHONPATH=. python3 -m py.cmd.icalc py/cmd/icalc/settings.json >output.txt
