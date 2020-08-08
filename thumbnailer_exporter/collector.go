package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metricCollector struct {
	upload          *prometheus.Desc
	uploadSpeed     *prometheus.Desc
	download        *prometheus.Desc
	downloadSpeed   *prometheus.Desc
	hashConsistency *prometheus.Desc
	imgfilesize     *prometheus.Desc
}

type metricOutput struct {
	collector string
	value     float64
}

func newmetricCollector() *metricCollector {
	return &metricCollector{
		//Constructing new metric label /w some description
		upload: prometheus.NewDesc("thumbs_upload_status",
			"Status of uploading image to thumbnailer, 0 ok;1 timed out or error; 2 failed",
			[]string{}, nil,
		),
		uploadSpeed: prometheus.NewDesc("thumbs_upload_speed",
			"Time it took to handle upload request in ms",
			[]string{}, nil,
		),
		download: prometheus.NewDesc("thumbs_download_status",
			"Status of downloading image from thumbnailer, 0 ok;1 timed out ; 2 failed",
			[]string{}, nil,
		),
		downloadSpeed: prometheus.NewDesc("thumbs_download_speed",
			"Time it took to handle download request in ms",
			[]string{}, nil,
		),
		hashConsistency: prometheus.NewDesc("thumbs_hash_consistency",
			"Status of hash consistency, 0 hashes are identical; 1 hashes are different ;2 failed to get hash",
			[]string{}, nil,
		),
		imgfilesize: prometheus.NewDesc("thumbs_img_size_bytes",
			"Size of testing image",
			[]string{}, nil,
		),
	}
}

func (collector *metricCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- collector.upload
	ch <- collector.uploadSpeed
	ch <- collector.download
	ch <- collector.downloadSpeed
	ch <- collector.hashConsistency
	ch <- collector.imgfilesize
}

func (collector *metricCollector) Collect(ch chan<- prometheus.Metric) {

	// Open channel

	collectorChan := make(chan metricOutput)

	// Run function thumbnailerCheck in goroutine (parallel with current code) and wait for it to end. Loop breaks when the channel in goroutine closes

	go thumbnailerCheck(collectorChan)
	for {
		collectorChan, status := <-collectorChan
		if status == false {
			break
		}
		switch collectorChan.collector {
		case "upload":
			ch <- prometheus.MustNewConstMetric(collector.upload, prometheus.GaugeValue, collectorChan.value)
			break
		case "uploadSpeed":
			ch <- prometheus.MustNewConstMetric(collector.uploadSpeed, prometheus.GaugeValue, collectorChan.value)
			break
		case "download":
			ch <- prometheus.MustNewConstMetric(collector.download, prometheus.GaugeValue, collectorChan.value)
			break
		case "downloadSpeed":
			ch <- prometheus.MustNewConstMetric(collector.downloadSpeed, prometheus.GaugeValue, collectorChan.value)
			break
		case "hashConsistency":
			ch <- prometheus.MustNewConstMetric(collector.hashConsistency, prometheus.GaugeValue, collectorChan.value)
			break
		case "imgfilesize":
			ch <- prometheus.MustNewConstMetric(collector.imgfilesize, prometheus.GaugeValue, collectorChan.value)
			break
		}
	}

}
