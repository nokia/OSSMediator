/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"bytes"
	"crypto/tls"
	"elasticsearchplugin/pkg/config"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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

	//indexList from which old data will be deleted
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

	indexMetaData = map[string]string{
		nhgData:         "nhg-data",
		simsData:        "sims-data",
		accountSimsData: "account-sims-data",
		apSimsData:      "ap-sims-data",
	}
)

const (
	//Timeout duration for HTTP calls
	timeout       = 60 * time.Second
	retryDuration = 5 * time.Minute

	elkBulkAPI           = "/_bulk"
	elkDeleteAPI         = "/_delete_by_query"
	elkCatIndicesAPI     = "/_cat/indices/"
	elkWaitQueryParam    = "wait_for_completion"
	elkIgnoreUnavailable = "ignore_unavailable"
	elkNoOfRecordsPerAPI = 500

	maxRetryAttempts = 3

	deletionHour    = 1 //Hour of the day, when deletion of old data will be done from elasticsearch
	timestampFormat = "2006-01-02T15:04:05"

	//Data types
	fmData          = "fmdata"
	pmData          = "pmdata"
	nhgData         = "network-hardware-groups"
	simsData        = "sims"
	accountSimsData = "account-sims"
	apSimsData      = "access-point-sims"
)

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

func PushData(filePath string, esConf config.ElasticsearchConf) {
	fileName := path.Base(filePath)
	apiType := strings.Split(fileName, "_")[0]
	if apiType == fmData {
		pushFMData(filePath, esConf)
	} else if apiType == pmData {
		pushPMData(filePath, esConf)
	} else if apiType == nhgData {
		pushNHGData(filePath, esConf)
	} else if apiType == simsData || apiType == accountSimsData {
		pushSimsData(filePath, esConf)
	} else if apiType == apSimsData {
		pushAPSimsData(filePath, esConf)
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

//PushFailedData pushes previously failed files from failedData map to elasticsearch every 5 min.
func PushFailedData(esConf config.ElasticsearchConf) {
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

func readFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading file: %s", filePath)
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading file: %s", filePath)
		return nil, err
	}
	return data, err
}
