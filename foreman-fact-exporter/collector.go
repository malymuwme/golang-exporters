package main

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type foremanCollectorStruct struct {
	stackconf *prometheus.Desc
}

type foremanOutput struct {
	metrictype string
	hostname   string
	value      float64
}

func newForemanCollector() *foremanCollectorStruct {
	return &foremanCollectorStruct{
		stackconf: prometheus.NewDesc("foreman_stackconf_runtime",
			"Shows for how long stackconf run",
			[]string{"metrictype", "hostname"}, nil,
		),
	}
}

func (collector *foremanCollectorStruct) Describe(ch chan<- *prometheus.Desc) {

	ch <- collector.stackconf
}

func (collector *foremanCollectorStruct) Collect(ch chan<- prometheus.Metric) {

	// This gets called on every request on /metrics . Here we create a channel to report metrics to and run background procs

	foremanChan := make(chan foremanOutput)
	log.Debug("run collectData")
	go collectData(foremanChan)
	for {

		// Infinite cycle waiting for channel foremanChan to be closed, so we can send the captured metrics to frontend

		foremanChan, status := <-foremanChan
		if status == false {
			break
		}
		log.Debug("********************************")
		log.Debug(status)
		log.Debug("Exported metrics")
		log.Debug(foremanChan.metrictype)
		log.Debug(foremanChan.value)
		log.Debug(foremanChan.hostname)
		ch <- prometheus.MustNewConstMetric(collector.stackconf, prometheus.GaugeValue, foremanChan.value, foremanChan.metrictype, foremanChan.hostname)
		log.Debug("********************************")
	}

}
