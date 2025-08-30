#!/usr/bin/env bash
set -euo pipefail

# --- Config ---
PLUGIN_ID="grafana-catalyst-datasource"

# Prefer git tag (vX.Y.Z) if present, else use package.json version
if git describe --tags --abbrev=0 >/dev/null 2>&1; then
  TAG=$(git describe --tags --abbrev=0)
  PLUGIN_VERSION="${TAG#v}"
else
  # requires npm 8+: prints quoted version, so trim quotes
  PLUGIN_VERSION=$(npm pkg get version | tr -d '"')
fi

ZIPNAME="${PLUGIN_ID}-${PLUGIN_VERSION}.zip"
TODAY=$(date +%Y-%m-%d)

echo "==> Building ${PLUGIN_ID} v${PLUGIN_VERSION}"

# --- Clean ---
rm -rf dist release "${ZIPNAME}"
mkdir -p dist

# --- Frontend build ---
npm run build

# NOTE: DO NOT overwrite dist/plugin.json with src/plugin.json.
# The Grafana build typically prepared dist/plugin.json already.

# --- Ensure plugin.json has version/date (replace placeholders if any) ---
if grep -q '%VERSION%' dist/plugin.json || grep -q '%TODAY%' dist/plugin.json; then
  echo "Patching plugin.json placeholders..."
  # portable sed (GNU/BSD)
  sed -i.bak "s/%VERSION%/${PLUGIN_VERSION}/g; s/%TODAY%/${TODAY}/g" dist/plugin.json || true
  rm -f dist/plugin.json.bak
fi

# --- Backend build (Linux AMD64) ---
GOOS=linux GOARCH=amd64 go build -o "dist/${PLUGIN_ID}_linux_amd64" ./cmd/grafana-catalyst-datasource
chmod +x "dist/${PLUGIN_ID}_linux_amd64"

# --- (Optional) include docs ---
# cp -f README.md CHANGELOG.md LICENSE dist/ 2>/dev/null || true

# --- Stage in release folder under top-level plugin dir ---
rm -rf release
mkdir -p "release/${PLUGIN_ID}"
cp -r dist/* "release/${PLUGIN_ID}/"

# Strip extended attributes to avoid upload parser issues
(
  cd release
  zip -r -X "../${ZIPNAME}" "${PLUGIN_ID}"
)

echo "==> Created release package: ${ZIPNAME}"
