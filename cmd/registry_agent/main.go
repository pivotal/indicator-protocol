package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/tls_config"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	intervalTime := flag.Duration("interval", 5*time.Minute, "The send interval")
	documentsGlob := flag.String("documents-glob", "/var/vcap/jobs/config/*/indicators.yml", "Glob path of indicator files")

	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Client TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")
	flag.Parse()

	startMetricsEndpoint()

	tlsConfig, err := tls_config.NewClientConfig(*clientPEM, *clientKey, *rootCACert, *serverCommonName)
	if err != nil {
		log.Fatal("Could not create registry agent, failed to create mTLS HTTP client")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   tlsConfig,
		},
	}

	agent := registry.Agent{
		DocumentFinder: registry.DocumentFinder{Glob: *documentsGlob},
		RegistryURI:    *registryURI,
		IntervalTime:   *intervalTime,
		Client:         client,
	}
	agent.Start()
}

func startMetricsEndpoint() {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Printf("unable to start monitor endpoint: %s", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	log.Printf("starting monitor endpoint on http://%s/metrics\n", lis.Addr().String())
	go func() {
		err = http.Serve(lis, mux)
		log.Printf("error starting the monitor server: %s", err)
	}()
}
