package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "dns_metrics"
)

var args struct {
	Port       int      `arg:"env:PORT, -p, --port" help:"port to listen on" default:"8080"`
	Timeout    int      `arg:"env:TIMEOUT, -t, --timeout" help:"timeout limit for lookups, in seconds" default:"5"`
	Interval   int      `arg:"env:INTERVAL, -i, --interval" help:"interval in minutes to perform lookup" default:"15"`
	Hostname   string   `arg:"env:LOOKUP_HOSTNAME, -n, --name" help:"the hostname to use for performing dns lookups" default:"google.com"`
	DNSServers []string `arg:"required, env:DNS_ADDRESS, -d, --dns-server" help:"dns server(s) to use for lookup, space separated"`
}

var (
	logger     = slog.Default()
	errorCount uint64
	hostname   string

	dnsDurationGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "dns_lookup_speed",
	}, []string{"host", "url", "dns_server"})
	dnsErrorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "dns_errors_total",
	}, []string{"host", "url", "dns_server"})
)

func main() {
	arg.MustParse(&args)
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		logger.Error("failed to get hostname", "error", err.Error())
		os.Exit(1)
	}
	go collectMetrics()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})
	http.Handle("/metrics", promhttp.Handler())
	logger.Info("server started", "interval", args.Interval, "port", args.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", args.Port), nil); err != nil {
		logger.Error("server stopped", "error", err)
	}
}

func collectMetrics() {
	for {
		for _, server := range args.DNSServers {
			duration, err := getDnsLookupTime(server)
			if err != nil {
				dnsErrorCounter.WithLabelValues(hostname, args.Hostname, server).Inc()
				logger.Error("failed to perform DNS lookup", "error", err.Error())
				continue
			}
			dnsDurationGauge.WithLabelValues(hostname, args.Hostname, server).Set(float64(duration))
		}
		time.Sleep(time.Duration(args.Interval) * time.Minute)
	}
}

func getDnsLookupTime(dnsAddress string) (time.Duration, error) {
	resolver := createResolver(dnsAddress)
	start := time.Now()
	_, err := resolver.LookupIP(context.Background(), "ip", args.Hostname)
	if err != nil {
		return time.Duration(0), fmt.Errorf("failed to lookup IP. %w", err)
	}
	elapsed := time.Since(start)
	return elapsed, nil
}

func createResolver(dnsServer string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Duration(args.Timeout) * time.Second,
			}
			return d.DialContext(ctx, network, dnsServer)
		},
	}
}
