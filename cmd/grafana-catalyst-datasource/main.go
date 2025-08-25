package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	log "github.com/grafana/grafana-plugin-sdk-go/backend/log"

	ds "github.com/extkljajicm/grafana-catalyst-datasource/pkg/backend"
)

func main() {
	log.DefaultLogger.Info("starting grafana-catalyst-datasource backend")

	d := ds.NewDatasource()

	if err := backend.Serve(backend.ServeOpts{
		CheckHealthHandler:  d,
		QueryDataHandler:    d,
		CallResourceHandler: d,
	}); err != nil {
		log.DefaultLogger.Error("failed to start plugin", "err", err)
	}
}
