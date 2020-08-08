package main

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/pagination"   
	log "github.com/sirupsen/logrus"
)

func rbd(krg string, cfg string, pool string, id string, name string, ch chan CephVolOut, projectName string) error {
	// Preparing ID for RBD DU
	imageID := pool + "/volume-" + id
	rbdOutput, err := exec.Command("/usr/bin/rbd", "du", imageID, "-c", cfg, "--keyring", krg).Output()
	if err == nil {
		reg, err := regexp.Compile("[^0-9.]+")
		if err != nil {
			log.Fatal(err)
		}
		pole := strings.Fields(string(rbdOutput))
		for i := 3; i < len(pole); i++ {
			if strings.Contains(pole[i], "@") {
				log.Debug("snapshot ", pole[i+2])
				log.Debug(pole[i])
				log.Debug(pole[i+1])
				log.Debug(reg.ReplaceAllString(pole[i+2], ""))
				tempvar, _ := strconv.ParseFloat(reg.ReplaceAllString(pole[i+2], ""), 64)
				if strings.Contains(pole[i+2], "MiB") {
					ch <- CephVolOut{tempvar * 1024, pole[i], "snapshot", name, projectName}
				} else if strings.Contains(pole[i+2], "GiB") {
					ch <- CephVolOut{tempvar * 1024 * 1024, pole[i], "snapshot", name, projectName}
				} else if strings.Contains(pole[i+2], "TiB") {
					ch <- CephVolOut{tempvar * 1024 * 1024 * 1024, pole[i], "snapshot", name, projectName}
				} else if strings.Contains(pole[i+2], "MB") {
					ch <- CephVolOut{tempvar * 1000, pole[i], "snapshot", name, projectName}
				} else if strings.Contains(pole[i+2], "GB") {
					ch <- CephVolOut{tempvar * 1000 * 1000, pole[i], "snapshot", name, projectName}
				} else if strings.Contains(pole[i+2], "TB") {
					ch <- CephVolOut{tempvar * 1000 * 1000 * 1000, pole[i], "snapshot", name, projectName}
				}
				i = i + 2
			} else if strings.Contains(pole[i], "<TOTAL>") {
				i = i + 2
			} else {
				log.Debug("volume ", pole[i+2])
				log.Debug(pole[i])
				log.Debug(pole[i+1])
				log.Debug(reg.ReplaceAllString(pole[i+2], ""))
				tempvar, _ := strconv.ParseFloat(reg.ReplaceAllString(pole[i+2], ""), 64)
				if strings.Contains(pole[i+2], "MiB") {
					ch <- CephVolOut{tempvar * 1024, pole[i], "volume", name, projectName}
				} else if strings.Contains(pole[i+2], "GiB") {
					ch <- CephVolOut{tempvar * 1024 * 1024, pole[i], "volume", name, projectName}
				} else if strings.Contains(pole[i+2], "TiB") {
					ch <- CephVolOut{tempvar * 1024 * 1024 * 1024, pole[i], "volume", name, projectName}
				} else if strings.Contains(pole[i+2], "MB") {
					ch <- CephVolOut{tempvar * 1000, pole[i], "volume", name, projectName}
				} else if strings.Contains(pole[i+2], "GB") {
					ch <- CephVolOut{tempvar * 1000, pole[i], "volume", name, projectName}
				} else if strings.Contains(pole[i+2], "TB") {
					ch <- CephVolOut{tempvar * 1000, pole[i], "volume", name, projectName}
				}
				i = i + 2
			}
		}
	} else {
		log.Info("unable to execute  rbd du  ", imageID, " , image is broken/in error state ")
	}
	return nil
}

func listVols(krg string, cfg string, pool string, ch chan CephVolOut, bashrc string) error {
	// Open and export env vars from openstackrc file
	rcfile, _ := os.Open(bashrc)
	rcscanner := bufio.NewScanner(rcfile)
	volreg, err := regexp.Compile("\"")
	if err != nil {
		log.Fatal(err)
	}
	for rcscanner.Scan() {
		line := volreg.ReplaceAllString(strings.TrimPrefix(rcscanner.Text(), "export "), "")
		if strings.Contains(line, "=") {
			ttt := strings.Split(line, "=")
			os.Setenv(ttt[0], ttt[1])
			log.Debug(os.Getenv(ttt[0]))
		}
	}
	if os.Getenv("OS_DOMAIN_NAME") == "" {
		os.Setenv("OS_DOMAIN_NAME", "Default")
	}
	if os.Getenv("OS_PROJECT_ID") == "" {
		os.Setenv("OS_PROJECT_ID", "default")
	}
	if os.Getenv("OS_PROJECT_NAME") == "" {
		os.Setenv("OS_PROJECT_NAME", "admin")
	}
	if os.Getenv("OS_INTERFACE") == "" {
		os.Setenv("OS_INTERFACE", "public")
	}
	// Parsing auth params from Env
	authopts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	// Auth func for Block Storage
	providerProjects, err := openstack.AuthenticatedClient(authopts)
	if err != nil {
		log.Info(providerProjects)
		log.Fatal(err)
	}
	// Auth func for Block Storage
	providerBlockStrg, err := openstack.AuthenticatedClient(authopts)
	if err != nil {
		log.Info(providerBlockStrg)
		log.Fatal(err)
	}

	//IDV3 Auth for project listing
	clientIDV3, err := openstack.NewIdentityV3(providerProjects, gophercloud.EndpointOpts{
		Region: "RegionOne",
	})
	if err != nil {
		log.Fatal(err)
	}

	// BlockStorage auth
	clientBlockStorage, err := openstack.NewBlockStorageV3(providerBlockStrg, gophercloud.EndpointOpts{
		Region: "RegionOne",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Options defined by volumes.ListOpts struct
	listProjects := projects.List(clientIDV3, projects.ListOpts{})

	var imageCount int

	_ = listProjects.EachPage(func(page pagination.Page) (bool, error) {
		projectPager, err := projects.ExtractProjects(page)
		log.Debug("staring project iteration")
		for _, projectPage := range projectPager {

			log.Debug("--------------------------------")
			log.Debug("Project ID ")
			log.Debug(projectPage.ID)
			log.Debug("Project Name ")
			log.Debug(projectPage.Name)
			log.Debug("--------------------------------")

			// Dyamic ptions for volume listing function
			opts := volumes.ListOpts{AllTenants: true, TenantID: projectPage.ID}

			// Getting output and parsing it into Pages, then processing the pages for data we need
			_ = volumes.List(clientBlockStorage, opts).EachPage(func(pagess pagination.Page) (bool, error) {
				volumePager, err := volumes.ExtractVolumes(pagess)
				for _, volumePage := range volumePager {

					imageCount++
					log.Debug("++++++++++++++++++++++++++++++++")
					log.Debug("volume ID - openstack ", volumePage.ID)

					// Check if volume type is ceph
					if strings.Contains(volumePage.VolumeType, "ceph") {
						rbd(krg, cfg, pool, volumePage.ID, volumePage.Name, ch, projectPage.Name)
						log.Debug("Processing image cycle : ", imageCount)
					}
					log.Debug("++++++++++++++++++++++++++++++++")
				}
				return true, err
			})
			log.Debug("Volume Listing Ended")
		}
		log.Debug("Project Listing Ended")
		return true, err
	})

	log.Debug("Done")
	log.Debug(imageCount)
	close(ch)
	return nil
}
