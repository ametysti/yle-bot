package prom

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	YleLatency = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "website_latency",
		Namespace: "yle_bot",
		Help:      "The total number of processed events",
	})

	BotLatency = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "websocket_latency",
		Namespace: "yle_bot",
		Help:      "Discord bot latency",
	})
)

func StartHTTPServer() {
	host := "localhost:3000"

	http.Handle("/metrics", promhttp.Handler())

	if _, err := os.Stat("/.dockerenv"); err == nil {
		host = ":3000"
	}

	http.ListenAndServe(host, nil)
}
