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
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	once        sync.Once
	netClient   *http.Client
	failedData  []failedResponse
	retryTicker *time.Ticker

	indexList   = []string{"edge-fm*", "edge-pm-*", "edge-ne3s-pm-*", "edge-ne3s-5g-pm*"}
	deleteQuery = `
{
  "query": {
    "range" : {
        "@timestamp": {
            "lte": "TIMESTAMP"
          }
    }
  }
}`
	currentTime = time.Now
)

const (
	//Timeout duration for HTTP calls
	timeout       = 60 * time.Second
	retryDuration = 5 * time.Minute

	elkBulkAPI           = "/_bulk"
	elkDeleteAPI         = "/_delete_by_query"
	elkWaitQueryParam    = "wait_for_completion"
	elkNoOfRecordsPerAPI = 200

	maxRetryAttempts = 3

	deletionHour    = 1 //Hour of the day, when deletion of old data will be done from elasticsearch
	timestampFormat = "2006-01-02T15:04:05"
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
	err := doPost(elkURL, data, nil)
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

func doPost(elkURL string, data string, queryParams map[string]string) error {
	request, err := http.NewRequest("POST", elkURL, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	request.Close = true
	request.Header.Set("Content-Type", "application/json")

	if len(queryParams) > 0 {
		query := request.URL.Query()
		for k, v := range queryParams {
			query.Add(k, v)
		}
		request.URL.RawQuery = query.Encode()
	}

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
		err = doPost(elkURL, data, nil)
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
				err := doPost(elkURL, data, nil)
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

//DeleteDataFormElasticsearch deletes older data from elasticsearch every day at 1 o'clock.
func DeleteDataFormElasticsearch(elkURL string, duration int) {
	timer := time.NewTimer(getNextTickDuration())
	for {
		<-timer.C
		log.Info("Triggered data cleanup of elasticsearch")
		deleteData(elkURL, duration)
		timer.Reset(getNextTickDuration())
	}
}

func getNextTickDuration() time.Duration {
	currTime := currentTime()
	nextTick := time.Date(currTime.Year(), currTime.Month(), currTime.Day(), deletionHour, 0, 0, 0, time.Local)
	if nextTick.Before(currTime) {
		nextTick = nextTick.AddDate(0, 0, 1)
	}
	return nextTick.Sub(currentTime())
}

func deleteData(elkURL string, duration int) {
	elkURL = elkURL + "/" + strings.Join(indexList, ",") + elkDeleteAPI
	deletionTime := currentTime().AddDate(0, 0, -1*duration).UTC().Format(timestampFormat)
	query := strings.Replace(deleteQuery, "TIMESTAMP", deletionTime, -1)
	queryParams := make(map[string]string)
	queryParams[elkWaitQueryParam] = "false"
	err := doPost(elkURL, query, queryParams)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "url": elkURL, "delete_from_time": deletionTime}).Error("Unable to delete data from elasticsearch")
	} else {
		log.WithFields(log.Fields{"url": elkURL, "delete_from_time": deletionTime}).Info("Data deleted from elasticsearch successfully.")
	}
}
