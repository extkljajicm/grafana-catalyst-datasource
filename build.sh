#!/usr/bin/env bash
# build.sh - Build both frontend and backend for grafana-catalyst-datasource
set -euo pipefail

echo "==> 1/2 Building frontend (Webpack)..."
npm run build

echo "==> 2/2 Building backend (Go/Mage)..."
go run github.com/magefile/mage -v BuildAll

echo "==> Build complete! Output located in ./dist"
ls -lh dist/
