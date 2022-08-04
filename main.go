package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	apiFormat = "%s/api/v1/%s"

	instanceHealthURI = "instance"
	instanceHealth    = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "instance_health",
		Help: "If /instance endpoint reports a healthy response.",
	})

	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total requests by exporter to mastodon instance.",
	}, []string{"status", "endpoint"})
)

func serveMetrics(errorChan chan<- error) {
	registry := prometheus.NewRegistry()
	reg := prometheus.WrapRegistererWithPrefix("mast_", registry)
	reg.MustRegister(instanceHealth)
	reg.MustRegister(requestsTotal)

	registry.Register(collectors.NewGoCollector())
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Printf("Serve the metrics http server.")
	errorChan <- http.ListenAndServe(":2112", nil)
}

func run(c *Config, errorChan chan<- error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(apiFormat, c.InstanceURI, instanceHealthURI), nil)
	if err != nil {
		errorChan <- fmt.Errorf("Unable to form request: %w", err)
	}
	req.Header.Set("User-Agent", "mastodon_exporter")

	ticker := time.NewTicker(c.CheckInterval)
	log.Printf("Start iterating on endpoints.")
	for range ticker.C {
		var reason string

		res, err := client.Do(req)

		requestsTotal.With(prometheus.Labels{
			"status":   fmt.Sprintf("%d", res.StatusCode),
			"endpoint": instanceHealthURI,
		}).Inc()

		switch {
		case err != nil:
			reason = err.Error()
		case res.StatusCode != http.StatusOK:
			reason = fmt.Sprintf("Response status: %s", res.Status)
		default:
			instanceHealth.Set(1)
			continue
		}
		instanceHealth.Set(0)
		c.StderrLogger.Printf("Endpoint %s failed with: %s\n", instanceHealthURI, reason)
	}
}

func main() {
	c := GenConfig()
	errChan := make(chan error)
	go serveMetrics(errChan)
	go run(&c, errChan)

	if err := <-errChan; err != nil {
		log.Fatal(err)
	}

}
