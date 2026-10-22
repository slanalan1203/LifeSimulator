package monitoring

import (
	"lifeSimulator/internal/colony"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewHandler(c *colony.Colony) http.Handler {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		newVillageCollector(c),
	)

	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
