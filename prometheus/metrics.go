package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	YleLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "website_latency",
		Namespace: "yle_bot",
		Help:      "The total number of processed events",
	})
)
