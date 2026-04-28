#!/usr/bin/env bash

cd "$(dirname "$0")/../../../" && zig build ifres -- src/cmd/ifres/settings.json >src/cmd/ifres/output.txt
