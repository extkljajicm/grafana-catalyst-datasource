PLUGIN_ID="grafana-catalyst-datasource"
PLUGIN_VERSION="0.1.0"
ZIPNAME="${PLUGIN_ID}-${PLUGIN_VERSION}.zip"

rm -rf dist/*
rm -rf release/*
npm run build
GOOS=linux GOARCH=amd64 go build -o dist/grafana-catalyst-datasource_linux_amd64 ./cmd/grafana-catalyst-datasource
chmod +x dist/grafana-catalyst-datasource_linux_amd64
cp src/plugin.json dist/

mkdir -p "release/${PLUGIN_ID}"
cp -r dist/* release/${PLUGIN_ID}/
(cd release && zip -r "../${ZIPNAME}" "${PLUGIN_ID}")

echo "Created release package: ${ZIPNAME}"
