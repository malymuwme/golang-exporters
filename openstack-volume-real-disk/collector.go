package main

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Prometheus metric descriptor
var (
	cephVolRealDisk = prometheus.NewDesc(
		prometheus.BuildFQName("ceph", "volume", "disk_usage"),
		"Amount of disk space a volume uses",
		[]string{"volumeId", "volumeName", "snapshot", "projectName"},
		nil)
)

type CephVolumes struct {
	pool    string
	config  string
	keyring string
	rcfile  string
}

type CephVolOut struct {
	value       float64
	id          string
	snap        string
	name        string
	projectname string
}

func CallExporter(pool string, config string, keyring string, bashrcfile string) (*CephVolumes, error) {
	return &CephVolumes{
		pool:    pool,
		config:  config,
		keyring: keyring,
		rcfile:  bashrcfile,
	}, nil
}

// Function triggered by CephVolumes

func (e *CephVolumes) Describe(ch chan<- *prometheus.Desc) {

	ch <- cephVolRealDisk
}

// Filling CephVolumes struct, the rest will be triggered as they are related to this struct

func CollectCephStuff(ch chan<- prometheus.Metric, value float64, volumeid string, snapshot string, name string, projectname string) error {

	ch <- prometheus.MustNewConstMetric(
		cephVolRealDisk,
		prometheus.GaugeValue,
		value,
		volumeid,
		name,
		snapshot,
		projectname)

	return nil
}


// Function triggered by CephVolumes , takes care about the functions which get values and then sending them into prometheus handler

func (e *CephVolumes) Collect(ch chan<- prometheus.Metric) {
	Cephch := make(chan CephVolOut)
	// Running listVols func in parallel, to get values from, then parsing it below in cycle untill the channel we just made was closed by the function
	log.Debug(" go listVols")
	go listVols(e.keyring, e.config, e.pool, Cephch, e.rcfile)

	for {
		Cephch, status := <-Cephch
		if status == false {
			// breaks loop if the channel closes
			break
		}
		log.Debug("********************************")
		log.Debug("Exported metrics")
		log.Debug("ID ", Cephch.id, "  NAME ", Cephch.name, "   VALUE ", Cephch.value, "  projectname   ", Cephch.projectname)
		CollectCephStuff(ch, float64(Cephch.value), Cephch.id, Cephch.snap, Cephch.name, Cephch.projectname)
		log.Debug("********************************")
	}
}

