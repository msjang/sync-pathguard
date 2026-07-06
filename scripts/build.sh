#!/usr/bin/env bash
# Local build for the host OS/arch.
#
# NOTE: the tray app uses cgo (fyne.io/systray → Cocoa/GTK/Win32), so the GUI
# cannot be cross-compiled to another OS from here. CI builds each OS natively
# (see .github/workflows). The scan/config engine is pure Go and cross-compiles
# fine if you ever need just the library.
set -euo pipefail
cd "$(dirname "$0")/.."

mkdir -p bin
out="bin/sync-pathguard"
[ "$(go env GOOS)" = "windows" ] && out="${out}.exe"

echo "Building Sync Pathguard for $(go env GOOS)/$(go env GOARCH)…"
go build -o "$out" ./cmd/sync-pathguard
echo "→ $out"
