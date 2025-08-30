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

# plugin.json: patch placeholders in dist (or copy from src if build didn't emit it)
if [ -f dist/plugin.json ]; then
  :
else
  cp src/plugin.json dist/
fi
if grep -q '%VERSION%' dist/plugin.json || grep -q '%TODAY%' dist/plugin.json; then
  sed -i.bak "s/%VERSION%/${PLUGIN_VERSION}/g; s/%TODAY%/${TODAY}/g" dist/plugin.json || true
  rm -f dist/plugin.json.bak
fi

# ---- Backend builds ----
PKG_PATH="./cmd/grafana-catalyst-datasource"

# (A) Primary runtime target for your container: linux/amd64 at legacy path
echo "-> linux/amd64 (legacy runtime path)"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w" -o "dist/${PLUGIN_ID}_linux_amd64" "${PKG_PATH}"
chmod +x "dist/${PLUGIN_ID}_linux_amd64"

# (B) Mirrors for validator (optional but helpful)
build_mirror () {
  local GOOS="$1" GOARCH="$2"
  local OUTDIR="dist/${GOOS}-${GOARCH}"
  mkdir -p "${OUTDIR}"
  local OUTBIN="${OUTDIR}/${PLUGIN_ID}"
  [ "${GOOS}" = "windows" ] && OUTBIN="${OUTBIN}.exe"
  echo "-> ${GOOS}/${GOARCH} (validator mirror)"
  GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o "${OUTBIN}" "${PKG_PATH}"
  chmod +x "${OUTBIN}" 2>/dev/null || true
}

build_mirror linux amd64
build_mirror linux arm64
build_mirror darwin amd64
build_mirror darwin arm64
build_mirror windows amd64

# Stage for zip
mkdir -p "release/${PLUGIN_ID}"
cp -R dist/* "release/${PLUGIN_ID}/"
# Ensure plugin.json at root of plugin dir as well
cp dist/plugin.json "release/${PLUGIN_ID}/plugin.json"

# Zip (strip xattrs)
( cd release && zip -r -X "../${ZIPNAME}" "${PLUGIN_ID}" )

echo "==> Created release package: ${ZIPNAME}"
