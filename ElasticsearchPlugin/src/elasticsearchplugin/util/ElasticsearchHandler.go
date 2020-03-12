/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	once        sync.Once
	netClient   *http.Client
	failedData  []failedResponse
	retryTicker *time.Ticker
)

const (
	//Timeout duration for HTTP calls
	timeout       = 60 * time.Second
	retryDuration = 5 * time.Minute

	elkBulkAPI           = "/_bulk"
	elkNoOfRecordsPerAPI = 200

	maxRetryAttempts = 3
)

type response []struct {
	Index  string      `json:"_index"`
	ID     string      `json:"_id"`
	Source interface{} `json:"_source"`
}

type failedResponse struct {
	filePath string
	data     string
}

func newNetClient() *http.Client {
	once.Do(func() {
		netTransport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		netClient = &http.Client{
			Transport: netTransport,
			Timeout:   timeout,
		}
	})

	return netClient
}

func pushDataToElasticsearch(filePath string, URL string) {
	elkURL := URL + elkBulkAPI
	log.Infof("Pushing data from %s to elasticsearch", filePath)
	f, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading file: %s", filePath)
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading file: %s", filePath)
		return
	}

	var resp response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}

	var postData string
	for i, d := range resp {
		source, _ := json.Marshal(d.Source)
		postData += `{"index": {"_index": "` + d.Index + `", "_id": "` + d.ID + `"}}` + "\n"
		postData += string(source) + "\n"
		if i != 0 && i%elkNoOfRecordsPerAPI == 0 || i == len(resp)-1 {
			pushData(elkURL, postData, filePath)
			postData = ""
		}
	}
}

func pushData(elkURL string, data string, filePath string) {
	err := doPost(elkURL, data)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "url": elkURL, "file": filePath}).Error("Unable to push data to elasticsearch, push to elasticsearch will be retried")
		err = retryPushData(elkURL, data)
		if err != nil {
			failedData = append(failedData, failedResponse{filePath: filePath, data: data})
			log.WithFields(log.Fields{"Error": err, "url": elkURL, "file": filePath}).Error("Unable to push data to elasticsearch, will be retried later...")
			return
		}
	}
	log.Infof("Data from %s pushed to elasticsearch successfully", filePath)
}

func doPost(elkURL string, data string) error {
	request, err := http.NewRequest("POST", elkURL, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	request.Close = true
	request.Header.Set("Content-Type", "application/json")
	response, err := newNetClient().Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Received response' status code: %d, status: %s", response.StatusCode, response.Status)
	}

	response.Body.Close()
	return nil
}

func retryPushData(elkURL string, data string) error {
	var err error
	for i := 0; i < maxRetryAttempts; i++ {
		log.WithFields(log.Fields{"url": elkURL}).Error("retrying push data to elasticsearch")
		err = doPost(elkURL, data)
		if err == nil {
			return nil
		}
	}
	return err
}

//PushFailedDataToElasticsearch pushes previously failed files from failedData map to elasticsearch every 5 min.
func PushFailedDataToElasticsearch(elkURL string) {
	elkURL = elkURL + elkBulkAPI
	if retryTicker == nil {
		retryTicker = time.NewTicker(retryDuration)
	}

	go func() {
		for {
			<-retryTicker.C
			for i := len(failedData) - 1; i >= 0; i-- {
				filePath := failedData[i].filePath
				data := failedData[i].data
				log.Infof("Retrying to push failed data from %s to elasticsearch", filePath)
				err := doPost(elkURL, data)
				if err == nil {
					failedData = append(failedData[:i], failedData[i+1:]...)
					log.Infof("Data from %s pushed to elasticsearch successfully", filePath)
				} else {
					log.WithFields(log.Fields{"Error": err, "url": elkURL, "file": filePath}).Error("Unable to push data to elasticsearch, push to elasticsearch will be retried")
				}
			}
		}
	}()
}
