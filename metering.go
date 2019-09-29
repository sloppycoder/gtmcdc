package gtmcdc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var counters map[string]prometheus.Counter
var histograms map[string]prometheus.Histogram

func IncrCounter(name string) {
	if counters == nil {
		counters = map[string]prometheus.Counter{}
	}

	counter, exists := counters[name]
	if !exists {
		counter = promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
		})
		counters[name] = counter
	}

	counter.Inc()
}

func HistoObserve(name string, value float64) {
	if histograms == nil {
		histograms = map[string]prometheus.Histogram{}
	}

	histo, exists := histograms[name]
	if !exists {
		histo = promauto.NewHistogram(prometheus.HistogramOpts{
			Name: name,
			// used to store microseconds
			Buckets: []float64{100, 250, 500, 1_000, 2_500, 5_000, 10_000, 25_000, 50_000, 100_000, 2_500_000},
		})
		histograms[name] = histo
	}
	histo.Observe(value)
}

func InitPromHttp(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Warn(http.ListenAndServe(addr, nil))
	}()
}