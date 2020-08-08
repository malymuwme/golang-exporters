package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/log"
)

func thumbnailerCheck(ch chan metricOutput) {
	file, err := os.Open(filename)
	if err != nil {
		log.Info("couldn't open the file " + filename)
		return
	}
	defer file.Close()

	// Get File size

	size, err := file.Stat()
	if err != nil {
		return
	}

	ch <- metricOutput{"imgfilesize", float64(size.Size())}

	startUpload := time.Now()

	// this part uploads the local file
	req, err := http.NewRequest("PUT", "http://"+url+":82/"+remotefilename, file)
	req.Header.Set("Content-Type", "image/jpeg")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		log.Info("There was an error in uploading the picture!", err)
		// Kill function when failed
		ch <- metricOutput{"upload", float64(2)}
		close(ch)
		return
	} else if res.StatusCode != 200 {
		// Kill function when failed
		ch <- metricOutput{"upload", float64(1)}
		close(ch)
		return
	}
	// Send UL data via channel made of metricOutput struct
	ch <- metricOutput{"upload", float64(0)}
	ch <- metricOutput{"uploadSpeed", float64(time.Since(startUpload).Milliseconds())}

	// time.Now().AddDate(0, 0, +1).Unix()
	timestamp := strconv.FormatInt(time.Now().Add(2*time.Hour).Unix(), 10)
	hashLink := []byte("/x" + remotefilename + ".200x200.jpg" + secret + timestamp)
	sumLink := md5.Sum(hashLink)
	base64Link := strings.Replace(strings.Replace(strings.Replace(base64.StdEncoding.EncodeToString(sumLink[:]), "/", "_", -1), "=", "", -1), "+", "-", -1)

	//
	// this part checks md5 of local and remote file
	//

	startDownload := time.Now()

	req0, err := http.NewRequest("GET", "http://"+url+"/x"+remotefilename+"."+imgsize+".jpg?vt="+timestamp+"&sg="+base64Link, nil)
	req.Header.Set("Content-Type", "image/jpeg")
	tr0 := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client0 := &http.Client{
		Transport: tr0,
		Timeout:   15 * time.Second,
	}
	res0, err := client0.Do(req0)
	if err != nil {
		log.Info("There was an error in getting the picture!", err)
		// Kill function when failed
		ch <- metricOutput{"download", float64(2)}
		close(ch)
		return
	}
	// Send DL data via channel made of metricOutput struct
	ch <- metricOutput{"download", float64(0)}
	ch <- metricOutput{"downloadSpeed", float64(time.Since(startDownload).Milliseconds())}

	// Check data from DL req with local

	bodyText, err := ioutil.ReadAll(res0.Body)
	r := bytes.NewReader(bodyText)
	hashRemote := calculateMD5(r)
	if err == nil {
		file, _ := os.Open(localfilename)
		defer file.Close()
		hashLocal := calculateMD5(file)
		if hashLocal == hashRemote {
			log.Info("Hashes of local and remote file are identical")
			ch <- metricOutput{"hashConsistency", float64(0)}
		} else {
			log.Info("Remote and local hash are not identical! Remote hash: ", hashRemote, "local hash: ", hashLocal)
			ch <- metricOutput{"hashConsistency", float64(1)}
		}
	} else {
		log.Info("error while parsing file for hash comparation.", err)
		ch <- metricOutput{"hashConsistency", float64(2)}
	}

	close(ch)

	//
}

func calculateMD5(reader io.Reader) string {
	/// this bit is partially taken from http://www.mrwaggel.be/post/generate-md5-hash-of-a-file-in-golang/ - thanks!
	var returnMD5String string
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		log.Info(returnMD5String, err)
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String
}
