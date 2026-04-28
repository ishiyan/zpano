#!/usr/bin/env bash

cd "$(dirname "$0")/../../../" && zig build iconf -- src/cmd/iconf/settings.json src/cmd/iconf/output
