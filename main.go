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

	errorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "Total errors encountered during runtime",
	})

	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total requests by exporter to mastodon instance.",
	}, []string{"status", "endpoint"})
)

func serveMetrics(c *Config, errorChan chan<- error) {
	registry := prometheus.NewRegistry()
	reg := prometheus.WrapRegistererWithPrefix("mast_", registry)
	reg.MustRegister(instanceHealth)
	reg.MustRegister(requestsTotal)

	if err := registry.Register(collectors.NewGoCollector()); err != nil {
		errorChan <- err
		return
	}
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Printf("Serve the metrics http server.")
	errorChan <- http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
}

func run(c *Config, errorChan chan<- error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(apiFormat, c.InstanceURI, instanceHealthURI), nil)
	if err != nil {
		errorChan <- fmt.Errorf("Unable to form request: %w", err)
		return
	}
	req.Header.Set("User-Agent", "mastodon_exporter")

	log.Printf("Start iterating on endpoints.")
	for range time.NewTicker(c.CheckInterval).C {
		if err := healthCheck(client, req); err != nil {
			c.StderrLogger.Println(err)
		}
	}
}

func healthCheck(client *http.Client, req *http.Request) error {
	res, err := client.Do(req)
	if err != nil {
		errorsTotal.Inc()
		return fmt.Errorf("Endpoint %s failed with: %w", instanceHealthURI, err)
	}

	defer res.Body.Close()

	requestsTotal.With(prometheus.Labels{
		"status":   fmt.Sprintf("%d", res.StatusCode),
		"endpoint": instanceHealthURI,
	}).Inc()

	if res.StatusCode != http.StatusOK {
		instanceHealth.Set(0)
		return fmt.Errorf("Endpoint %s failed with: Status code is not OK", instanceHealthURI)
	}

	instanceHealth.Set(1)
	return nil
}

func main() {
	c := GenConfig()
	errChan := make(chan error)
	go serveMetrics(&c, errChan)
	go run(&c, errChan)

	if err := <-errChan; err != nil {
		log.Fatal(err)
	}
}
