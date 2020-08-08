package main

import (
	"crypto/tls"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/tidwall/gjson"
)

func httpCall(url string) []byte {

	//Basic httpCall func using login, nthng special

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	curlclient := http.Client{
		Transport: tr,
		Timeout:   time.Second * 60,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "foreman-exporter")
	req.SetBasicAuth(*foremanUser, *foremanPassword)

	res, getErr := curlclient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	return body

}

func talkToAPI(url string, page int, wg *sync.WaitGroup, foreman string, ch chan foremanOutput) {

	defer wg.Done()



	log.Debug(url + "&page=" + strconv.Itoa(page))

	// List all hosts in foreman and read their name

	jsonOutput := httpCall(url + "&page=" + strconv.Itoa(page))

	listResult := gjson.Get(string(jsonOutput), "results.#.name")

	// For each hostname run another query for their facts

	for _, name := range listResult.Array() {
		log.Debug(name.String() + "   page " + strconv.Itoa(page))
		hostURL := foreman + "/" + name.String()
		hostResponse := httpCall(hostURL)
		
		// List the facts and iterate over them

		listResult := gjson.Get(string(hostResponse), "all_parameters")
		log.Debug("before listResult" + "   page " + strconv.Itoa(page))
		listResult.ForEach(func(key, value gjson.Result) bool {

			// If current fact is stackconf pappete runtime/ppt_runtime then parse it and send it to prometheus client via channel

			if gjson.Get(value.String(), "name").String() == "stackconf_puppet_runtime" {
				log.Debug("  -----------    stackconf_puppet_runtime     ------------" + "   page " + strconv.Itoa(page))
				log.Debug("sent" + "   page " + strconv.Itoa(page))
				log.Debug(gjson.Get(value.String(), "value").String() + "   page " + strconv.Itoa(page))
				if gjson.Get(value.String(), "value").Value() != nil {
					floatLen := len(strings.Split(gjson.Get(value.String(), "value").String(), ","))
					for i := 0; i < floatLen; i++ {
						float1, _ := strconv.ParseFloat(strings.Split(gjson.Get(value.String(), "value").String(), ",")[i], 32)
						ch <- foremanOutput{"puppet_runtime_" + strconv.Itoa(i+1), name.String(), float1}
					}
				}
				log.Debug("  ----------------------------------------" + "   page " + strconv.Itoa(page))
			}
			if gjson.Get(value.String(), "name").String() == "stackconf_runtime" {
				log.Debug("  -----------    stackconf_runtime     ------------" + "   page " + strconv.Itoa(page))
				log.Debug(gjson.Get(value.String(), "value").String() + "   page " + strconv.Itoa(page))
				if gjson.Get(value.String(), "value").Value() != nil {
					ch <- foremanOutput{"runtime", name.String(), gjson.Get(value.String(), "value").Float()}
				}

				log.Debug("  -----------------------" + "   page " + strconv.Itoa(page))
			}
			log.Debug("return true" + "   page " + strconv.Itoa(page))
			return true // keep iterating

		})
		log.Debug("after foreach" + "   page " + strconv.Itoa(page))

	}
	log.Debug("after for")

}

func collectData(ch chan foremanOutput) {

	// Get host count and split it into *thredcount threads to decrease time of run. Or to fry off foreman...

	foreman := "https://" + *foremanFlag + "/api/hosts"
	//callURL := foreman + "?per_page=5000&thin=true"

	getHostCount := httpCall(foreman)
	hostCount := gjson.Get(string(getHostCount), "total").Float()

	perPage := math.Ceil(hostCount / float64(*threadcount))

	// Create syngroup so the cycle wont overrun the exporter

	var wg sync.WaitGroup

	fullURL := foreman + "?per_page=" + strconv.FormatFloat(perPage, 'f', 0, 64) + "&thin=true"
	//fullURL := foreman + "?per_page=10" + "&thin=true"

	for i := 1; i < *threadcount+1; i++ {
		log.Debug("page ")
		log.Debug(i)
		wg.Add(1)
		go talkToAPI(fullURL, i, &wg, foreman, ch)
		log.Debug("ran goroutine")
	}
	wg.Wait()

	// Close channel when wg.Wait will finnish waiting for child procs

	close(ch)
	//log.Info("Finnished")
}
