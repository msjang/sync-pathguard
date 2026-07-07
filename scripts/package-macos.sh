#!/usr/bin/env bash
# Package the macOS build: a "Pathguard.app" menu-bar agent + the pathguard
# CLI, zipped for release. Builds NATIVELY for the host arch (systray is cgo, so
# CI runs this on both an arm64 and an amd64 runner — see .github/workflows).
#
# Usage: scripts/package-macos.sh [version] [arch-label]
set -euo pipefail
cd "$(dirname "$0")/.."

VER="${1:-0.0.0-dev}"
VER_NUM="${VER#v}"
ARCH="${2:-$(go env GOARCH)}"

APP="dist/Pathguard.app"
rm -rf dist
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

echo "Building tray app (cgo, $(go env GOARCH))…"
CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" \
	-o "$APP/Contents/MacOS/pathguard-gui" ./cmd/pathguard-gui

echo "Building CLI (pure Go)…"
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/pathguard ./cmd/pathguard

cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key><string>Pathguard</string>
	<key>CFBundleDisplayName</key><string>Pathguard</string>
	<key>CFBundleIdentifier</key><string>io.github.msjang.pathguard</string>
	<key>CFBundleExecutable</key><string>pathguard-gui</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>${VER_NUM}</string>
	<key>CFBundleVersion</key><string>${VER_NUM}</string>
	<key>LSMinimumSystemVersion</key><string>10.15</string>
	<key>LSUIElement</key><true/>
	<key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST

echo "Zipping…"
ditto -c -k --keepParent "$APP" "dist/Pathguard-macos-${ARCH}.zip"
( cd dist && zip -q "pathguard-macos-${ARCH}.zip" pathguard )

echo "→ dist/Pathguard-macos-${ARCH}.zip"
echo "→ dist/pathguard-macos-${ARCH}.zip"
