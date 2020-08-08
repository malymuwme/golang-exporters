package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var url string
var filename string
var localfilename string
var remotefilename string
var secret string
var exporterBindAddress string
var imgsize string

func init() {
	flag.StringVar(&url, "url", "", "URL to upload to/read from -without http")
	flag.StringVar(&filename, "filename", "", "File to upload")
	flag.StringVar(&localfilename, "localfilename", "", "File to check against")
	flag.StringVar(&remotefilename, "remotefilename", "monitoring2", "file to check on remote")
	flag.StringVar(&secret, "secret", "", "secret for md5 calc")
	flag.StringVar(&exporterBindAddress, "exporteraddr", ":9112", "Address on which to expose metrics and web interface.")
	flag.StringVar(&imgsize, "imgsize", "200x200", "x and y size of the testing img")
	flag.Parse()
}
func main() {

	if url == "" || filename == "" {
		fmt.Println("Couldn't parse all the necessary parameters. Please check the arguments")
		os.Exit(2)
	}

	//Create a new instance of the metricCollector and
	//register it with the prometheus client.
	foo := newmetricCollector()
	prometheus.MustRegister(foo)

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		  <head><title>Thumbnailer exporter</title></head>
		  <body>
		  <h1>Thumbnailer exporter</h1>
		  <p><a href="/metrics">Metrics</a></p>
		  </body>
		  </html>`))
	})
	log.Info("Beginning to serve on port " + exporterBindAddress)
	log.Fatal(http.ListenAndServe(exporterBindAddress, nil))

}
