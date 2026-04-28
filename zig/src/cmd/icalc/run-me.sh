#!/usr/bin/env bash

cd "$(dirname "$0")/../../../" && zig build icalc -- src/cmd/icalc/settings.json >src/cmd/icalc/output.txt
