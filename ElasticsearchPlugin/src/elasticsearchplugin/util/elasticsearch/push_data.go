/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"bytes"
	"crypto/tls"
	"elasticsearchplugin/config"
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

	indexList   = []string{"radio-fm*", "dac-fm*", "core-fm*", "4g-pm*", "5g-pm*"}
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
	fmData          = "fmdata"
	pmData          = "pmdata"
	nhgData         = "network-hardware-groups"
	simsData        = "sims"
	accountSimsData = "account-sims"
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

func PushDataToElasticsearch(filePath string, esConf config.ElasticsearchConf) {
	fileName := path.Base(filePath)
	apiType := strings.Split(fileName, "_")[0]
	if apiType == fmData {
		pushFMDataToElasticsearch(filePath, esConf)
	} else if apiType == pmData {
		pushPMDataToElasticsearch(filePath, esConf)
	} else if apiType == nhgData {
		pushNHGDataToElasticsearch(filePath, esConf)
	} else if apiType == simsData || apiType == accountSimsData {
		pushSimsDataToElasticsearch(filePath, esConf)
	}
}

func pushData(elkURL, elkUser, elkPassword string, data string, filePath string) {
	_, err := httpCall(http.MethodPost, elkURL, elkUser, elkPassword, data, nil)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "url": elkURL, "file": filePath}).Error("Unable to push data to elasticsearch, push to elasticsearch will be retried")
		err = retryPushData(elkURL, elkUser, elkPassword, data)
		if err != nil {
			failedData = append(failedData, failedResponse{filePath: filePath, data: data})
			log.WithFields(log.Fields{"Error": err, "url": elkURL, "file": filePath}).Error("Unable to push data to elasticsearch, will be retried later...")
			return
		}
	}
	log.Infof("Data from %s pushed to elasticsearch successfully", filePath)
}

func httpCall(httpMethod, elkURL, elkUser, elkPassword string, data string, queryParams map[string]string) ([]byte, error) {
	request, err := http.NewRequest(httpMethod, elkURL, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	request.Close = true
	if elkUser != "" && elkPassword != "" {
		request.SetBasicAuth(elkUser, elkPassword)
	}
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
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received response' status code: %d, status: %s", response.StatusCode, response.Status)
	}

	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func retryPushData(elkURL, elkUser, elkPassword string, data string) error {
	var err error
	for i := 0; i < maxRetryAttempts; i++ {
		log.WithFields(log.Fields{"url": elkURL}).Error("retrying push data to elasticsearch")
		_, err = httpCall(http.MethodPost, elkURL, elkUser, elkPassword, data, nil)
		if err == nil {
			return nil
		}
	}
	return err
}

//PushFailedDataToElasticsearch pushes previously failed files from failedData map to elasticsearch every 5 min.
func PushFailedDataToElasticsearch(esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
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
				_, err := httpCall(http.MethodPost, elkURL, esConf.User, esConf.Password, data, nil)
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

func pushFMDataToElasticsearch(filePath string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
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
			pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
			postData = ""
		}
	}
}

func pushPMDataToElasticsearch(filePath string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
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
			pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
			postData = ""
		}
	}
}

func pushNHGDataToElasticsearch(filePath string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
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
	pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
}

func pushSimsDataToElasticsearch(filePath string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
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
	index := metric + "-data"
	id := strings.Join([]string{metric, nhgID}, "_")
	postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
	postData += string(source) + "\n"
	pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
}
