package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Function for listing projects and their volumes (ID, Name ..) and then calling rbd function to get an actuall disk usage

func main() {
	var (
		cephkeyring = flag.String("keyring", "/etc/ceph/ceph.client.admin.keyring", "Keyring path")
		cephcfg     = flag.String("config", "/etc/ceph/ceph.conf", "Config path")
		cephpool    = flag.String("pool", "", "Ceph pool")
		port        = flag.String("port", ":9696", "Port number for exporter")
		bashrc      = flag.String("bashrc", "/root/keystonerc_admin", "bashrc path")
		usetls      = flag.Bool("usetls", false, "Flag for TLS usage")
		debuglog    = flag.Bool("debuglog", false, "Debug log level")
	)
	flag.Parse()

	if *debuglog == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	exporter, err := CallExporter(*cephpool, *cephcfg, *cephkeyring, *bashrc)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
    <head><title>Ceph Volumes DU exporter</title></head>
    <body>
    <h1>Ceph Volumes DU exporter</h1>
    <p><a href="/metrics">Metrics</a></p>
    </body>
    </html>`))
	})

	if *usetls {
		// Gets current path for .pem files
		execpath, _ := os.Executable()
		execfolder := filepath.Dir(execpath)

		log.Info("Use TLS flag : ", *usetls)
		// Create a CA certificate pool and add cert.pem to it
		caCert, err := ioutil.ReadFile(execfolder + "/cert.pem")
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Create the TLS Config with the CA pool and enable Client certificate validation
		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
		tlsConfig.BuildNameToCertificate()

		// Create a Server instance to listen on port 8443 with the TLS config
		server := &http.Server{
			Addr:      *port,
			TLSConfig: tlsConfig,
		}
		log.Info("Beginning to serve on port ", *port)
		log.Fatal(server.ListenAndServeTLS(execfolder+"/cert.pem", execfolder+"/key.pem"))
	} else {
		log.Info("Beginning to serve on port ", *port)
		log.Fatal(http.ListenAndServe(*port, nil))
	}
}
