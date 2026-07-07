#!/usr/bin/env bash
# Package the macOS build: a UNIVERSAL (arm64 + amd64) "Pathguard.app" menu-bar
# agent plus a universal pathguard CLI, zipped for release.
#
# The app is cgo (systray), so each arch is built with `clang -arch …` and
# combined with lipo. This runs fine on a single Apple-Silicon runner, so we
# don't depend on the scarce/slow Intel macOS runners.
#
# Usage: scripts/package-macos.sh [version]
set -euo pipefail
cd "$(dirname "$0")/.."

VER="${1:-0.0.0-dev}"
VER_NUM="${VER#v}"

APP="dist/Pathguard.app"
rm -rf dist build
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources" build

echo "Building universal tray app (cgo: arm64 + amd64)…"
CGO_ENABLED=1 GOARCH=arm64 CC="clang -arch arm64" \
	go build -trimpath -ldflags "-s -w" -o build/gui-arm64 ./cmd/pathguard-gui
CGO_ENABLED=1 GOARCH=amd64 CC="clang -arch x86_64" \
	go build -trimpath -ldflags "-s -w" -o build/gui-amd64 ./cmd/pathguard-gui
lipo -create -output "$APP/Contents/MacOS/pathguard-gui" build/gui-arm64 build/gui-amd64

echo "Building universal CLI (pure Go)…"
CGO_ENABLED=0 GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/cli-arm64 ./cmd/pathguard
CGO_ENABLED=0 GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/cli-amd64 ./cmd/pathguard
lipo -create -output dist/pathguard build/cli-arm64 build/cli-amd64

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
ditto -c -k --keepParent "$APP" "dist/Pathguard-macos-universal.zip"
( cd dist && zip -q "pathguard-macos-universal.zip" pathguard )

echo "→ dist/Pathguard-macos-universal.zip"
echo "→ dist/pathguard-macos-universal.zip"
