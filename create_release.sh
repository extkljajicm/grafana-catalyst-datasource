#!/usr/bin/env bash
set -euo pipefail

PLUGIN_ID="grafana-catalyst-datasource"

# Prefer git tag (vX.Y.Z) if present, else use package.json version
if git describe --tags --abbrev=0 >/dev/null 2>&1; then
  TAG=$(git describe --tags --abbrev=0)
  PLUGIN_VERSION="${TAG#v}"
else
  PLUGIN_VERSION=$(npm pkg get version | tr -d '"')
fi

ZIPNAME="${PLUGIN_ID}-${PLUGIN_VERSION}.zip"
TODAY=$(date +%Y-%m-%d)

echo "==> Building ${PLUGIN_ID} v${PLUGIN_VERSION}"

# Clean
rm -rf dist release "${ZIPNAME}"
mkdir -p dist

# Frontend
npm run build

# Ensure plugin.json in dist and patch placeholders if present
if [ ! -f dist/plugin.json ]; then
  cp src/plugin.json dist/
fi
if grep -q '%VERSION%' dist/plugin.json || grep -q '%TODAY%' dist/plugin.json; then
  sed -i.bak "s/%VERSION%/${PLUGIN_VERSION}/g; s/%TODAY%/${TODAY}/g" dist/plugin.json || true
  rm -f dist/plugin.json.bak
fi

# Backend (ONLY linux/amd64; legacy path expected by Grafana runtime)
PKG_PATH="./cmd/grafana-catalyst-datasource"
echo "-> linux/amd64 (runtime)"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w" -o "dist/${PLUGIN_ID}_linux_amd64" "${PKG_PATH}"
chmod +x "dist/${PLUGIN_ID}_linux_amd64"

# Optional: shrink frontend by stripping sourcemaps if present
find dist -type f -name "*.map" -delete || true

# Stage and zip
mkdir -p "release/${PLUGIN_ID}"
cp -R dist/* "release/${PLUGIN_ID}/"
cp dist/plugin.json "release/${PLUGIN_ID}/plugin.json"
cp -R src/img "release/${PLUGIN_ID}/"

( cd release && zip -r -X "../${ZIPNAME}" "${PLUGIN_ID}" )

echo "==> Created release package: ${ZIPNAME}"
