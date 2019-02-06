package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/pivotal/indicator-protocol/pkg/mtls"
	"github.com/pivotal/indicator-protocol/pkg/prometheus_alerts"
	"github.com/pivotal/indicator-protocol/pkg/registry"
)

func main() {
	registryURI := flag.String("registry", "", "URI of a registry instance")
	outputDirectory := flag.String("output-directory", "", "Indicator output-directory URI")
	clientPEM := flag.String("tls-pem-path", "", "Client TLS public cert pem path which can connect to the server (indicator-registry)")
	clientKey := flag.String("tls-key-path", "", "Server TLS private key path which can connect to the server (indicator-registry)")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs")
	serverCommonName := flag.String("tls-server-cn", "indicator-registry", "server (indicator-registry) common name")
	prometheusURI := flag.String("prometheus", "", "URI of a Prometheus server instance")

	flag.Parse()

	tlsConfig, err := mtls.NewClientConfig(*clientPEM, *clientKey, *rootCACert, *serverCommonName)
	if err != nil {
		log.Fatalf("failed to create mtls http client, %s", err)
	}

	registryHttpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   tlsConfig,
		},
	}

	prometheusHttpClient := &http.Client{}

	apiClient := registry.NewAPIClient(*registryURI, registryHttpClient)
	prometheusClient := prometheus_alerts.NewPrometheusClient(*prometheusURI, prometheusHttpClient)

	c := prometheus_alerts.ControllerConfig{
		RegistryAPIClient:   apiClient,
		PrometheusAPIClient: prometheusClient,
		OutputDirectory:     *outputDirectory,
	}

	controller := prometheus_alerts.NewController(c)
	err = controller.Update()
	if err != nil {
		log.Fatalf("failed to update prometheus alerts:, %s", err)
	}
}
