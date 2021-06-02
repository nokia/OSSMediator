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
	"path"
	"strconv"
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

	indexList   = []string{"radio-fm*", "dac-fm*", "4g-pm*", "5g-pm*"}
	deleteQuery = `
{
  "query": {
    "range" : {
        "timestamp": {
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
	elkCatIndicesAPI     = "/_cat/indices/"
	elkWaitQueryParam    = "wait_for_completion"
	elkNoOfRecordsPerAPI = 200

	maxRetryAttempts = 3

	deletionHour    = 1 //Hour of the day, when deletion of old data will be done from elasticsearch
	timestampFormat = "2006-01-02T15:04:05"

	//Data types
	fmData   = "fmdata"
	pmData   = "pmdata"
	nhgData  = "network-hardware-groups"
	simsData = "sims"
)

type pmResponse []struct {
	PMData       map[string]interface{} `json:"pm_data"`
	PMDataSource map[string]interface{} `json:"pm_data_source"`
	Timestamp    time.Time              `json:"timestamp"`
}

type fmResponse []struct {
	FMData       map[string]interface{} `json:"fm_data"`
	FMDataSource map[string]interface{} `json:"fm_data_source"`
	Timestamp    time.Time              `json:"timestamp"`
}

type failedResponse struct {
	filePath string
	data     string
}

type genericData struct {
	Data      string
	Timestamp time.Time `json:"timestamp"`
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
	fileName := path.Base(filePath)
	apiType := strings.Split(fileName, "_")[0]
	if apiType == fmData {
		pushFMDataToElasticsearch(filePath, URL)
	} else if apiType == pmData {
		pushPMDataToElasticsearch(filePath, URL)
	} else if apiType == nhgData {
		pushNHGDataToElasticsearch(filePath, URL)
	} else if apiType == simsData {
		pushSimsDataToElasticsearch(filePath, URL)
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

func pushFMDataToElasticsearch(filePath string, URL string) {
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

	fileName := path.Base(filePath)
	metricType := strings.Split(fileName, "_")[1]
	metricType = strings.ToLower(metricType)

	var resp fmResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}

	var postData string
	index := strings.Join([]string{metricType, "fm"}, "-")
	for i, d := range resp {
		dn := d.FMDataSource["dn"].(string)
		eventTime := d.FMData["event_time"].(string)
		hwID := d.FMDataSource["hw_id"].(string)

		var id string
		if metricType == "radio" {
			id = strings.Join([]string{hwID, eventTime, dn}, "_")
		} else if metricType == "dac" {
			alarmID := d.FMData["alarm_identifier"].(string)
			faultID := d.FMData["fault_id"].(string)
			keys := []string{metricType, hwID, dn, alarmID}
			if faultID != "" {
				keys = append(keys, faultID)
			}
			keys = append(keys, eventTime)
			id = strings.Join(keys, "_")
		} else if metricType == "core" {
			alarmID := d.FMData["alarm_identifier"].(string)
			nhgID := d.FMDataSource["nhg_id"].(string)
			edgeID := d.FMDataSource["edge_id"].(string)
			keys := []string{metricType, alarmID, nhgID, edgeID, eventTime}
			id = strings.Join(keys, "_")
		}
		d.Timestamp = time.Now()

		source, _ := json.Marshal(d)
		postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
		postData += string(source) + "\n"
		if i != 0 && i%elkNoOfRecordsPerAPI == 0 || i == len(resp)-1 {
			pushData(elkURL, postData, filePath)
			postData = ""
		}
	}
}

func pushPMDataToElasticsearch(filePath string, URL string) {
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

	var resp pmResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}

	var postData string
	currTime := time.Now().UTC()
	for i, d := range resp {
		hwID := d.PMDataSource["hw_id"].(string)
		dn := d.PMDataSource["dn"].(string)
		eventTime := d.PMDataSource["timestamp"].(string)
		technology := d.PMDataSource["technology"].(string)
		var objectType string
		for k := range d.PMData {
			objectType = k
			break
		}
		objectType = strings.ToLower(objectType[:strings.LastIndex(objectType, "_")])

		index := strings.Join([]string{strings.ToLower(technology), "pm", objectType, fmt.Sprintf("%02d", currTime.Month()), strconv.Itoa(currTime.Year())}, "-")
		id := strings.Join([]string{hwID, eventTime, dn}, "_")

		d.Timestamp = time.Now()
		source, _ := json.Marshal(d)
		postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
		postData += string(source) + "\n"
		if i != 0 && i%elkNoOfRecordsPerAPI == 0 || i == len(resp)-1 {
			pushData(elkURL, postData, filePath)
			postData = ""
		}
	}
}

func pushNHGDataToElasticsearch(filePath string, URL string) {
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

	fileName := path.Base(filePath)
	keys := strings.Split(fileName, "_")
	metric := strings.ToLower(keys[0])
	user := strings.ToLower(keys[1])

	d := struct {
		Data      interface{} `json:"network_info"`
		Timestamp time.Time   `json:"timestamp"`
	}{
		Timestamp: time.Now(),
	}
	err = json.Unmarshal(data, &d.Data)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}
	source, _ := json.Marshal(d)

	var postData string
	index := "nhg-details"
	id := strings.Join([]string{metric, user}, "_")
	postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
	postData += string(source) + "\n"
	pushData(elkURL, postData, filePath)
}

func pushSimsDataToElasticsearch(filePath string, URL string) {
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

	fileName := path.Base(filePath)
	keys := strings.Split(fileName, "_")
	metric := strings.ToLower(keys[0])
	nhgID := strings.ToLower(keys[1])

	d := struct {
		Data      interface{} `json:"sim_data"`
		Timestamp time.Time   `json:"timestamp"`
	}{
		Timestamp: time.Now(),
	}
	err = json.Unmarshal(data, &d.Data)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}
	source, _ := json.Marshal(d)

	var postData string
	index := "sims-data"
	id := strings.Join([]string{metric, nhgID}, "_")
	postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
	postData += string(source) + "\n"
	pushData(elkURL, postData, filePath)
}
