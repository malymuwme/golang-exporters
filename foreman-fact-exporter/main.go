package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	threadcount     = flag.Int("threadcount", 4, "Thread count in which it sends API requests")
	foremanFlag     = flag.String("foreman", "foreman1.infra.prod.ci", "Config path")
	foremanUser     = flag.String("foremanUser", "", "user for API")
	foremanPassword = flag.String("foremanPassword", "", "password for API")
	port            = flag.Int("port", 9119, "Port for exporter")
	loglevel        = flag.String("loglevel", "info", "Log level ; info / debug")
)

func main() {

	flag.Parse()

	if *loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	}
	//Create a new instance of the foremanCollectorcollector and
	//register it with the prometheus client.
	foremanCollector := newForemanCollector()
	prometheus.MustRegister(foremanCollector)

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		  <head><title>Foreman Stackconf exporter</title></head>
		  <body>
		  <h1>Foreman Stackconf exporter</h1>
		  <p><a href="/metrics">Metrics</a></p>
		  </body>
		  </html>`))
	})
	log.Info("Beginning to serve on port :" + strconv.Itoa(*port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
