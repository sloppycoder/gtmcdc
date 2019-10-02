package gtmcdc

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

var counters map[string]prometheus.Counter
var histograms map[string]prometheus.Histogram

// GetCounterValue returns the value of a counter
// A new counter will be created if it does not exist already
func GetCounterValue(name string) float64 {
	counter, exists := counters[name]
	if !exists {
		return 0
	}

	pb := &dto.Metric{}
	_ = counter.Write(pb)
	return pb.GetCounter().GetValue()
}

// IncrCounter increment a counter
// A new counter will be created if it does not exist already
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

// HistoObserve records an obseration for a histogram
// A new histogram will be created if it does not exist already
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

// InitPromHTTP starts the Http Listener that export the Prometheus metrics
// that can be scraped by Prometheus
func InitPromHTTP(addr string) error {
	var err error

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err = http.ListenAndServe(addr, nil)
	}()

	return err
}
