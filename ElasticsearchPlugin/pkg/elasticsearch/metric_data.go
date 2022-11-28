/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"elasticsearchplugin/pkg/config"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
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

const (
	corePMMapping = `{
  "mappings": {
    "dynamic_templates": [
      {
        "pm_data": {
          "match_mapping_type": "long",
          "mapping": {
            "type": "float"
          }
        }
      }
    ]
  }
}`
	reindexPostData = `{"source":{"index":"SOURCE"},"dest":{"index":"DEST"}}`
)

func pushFMData(filePath string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
	log.Infof("Pushing data from %s to elasticsearch", filePath)
	data, err := readFile(filePath)
	if err != nil {
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
		if d.FMData["event_time"] == nil {
			continue
		}
		dn := d.FMDataSource["dn"].(string)
		eventTime := d.FMData["event_time"].(string)
		hwID := d.FMDataSource["hw_id"].(string)
		alarmID := d.FMData["alarm_identifier"].(string)
		var id string
		if metricType == "radio" {
			specificProb := d.FMData["specific_problem"].(string)
			id = strings.Join([]string{hwID, dn, alarmID, specificProb, eventTime}, "_")
		} else {
			faultID := d.FMData["fault_id"].(string)
			keys := []string{metricType, hwID, dn, alarmID}
			if faultID != "" {
				keys = append(keys, faultID)
			}
			keys = append(keys, eventTime)
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

func pushPMData(filePath string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
	log.Infof("Pushing data from %s to elasticsearch", filePath)
	data, err := readFile(filePath)
	if err != nil {
		return
	}

	var resp pmResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}

	fileName := path.Base(filePath)
	metricType := strings.Split(fileName, "_")[1]
	metricType = strings.ToLower(metricType)

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

		var index, id string
		if metricType == "radio" {
			objectType = strings.ToLower(objectType[:strings.LastIndex(objectType, "_")])
			index = strings.Join([]string{strings.ToLower(technology), "pm", objectType, fmt.Sprintf("%02d", currTime.Month()), strconv.Itoa(currTime.Year())}, "-")
			id = strings.Join([]string{hwID, eventTime, dn}, "_")
		} else {
			index = strings.Join([]string{metricType, "pm"}, "-")
			id = strings.Join([]string{objectType, dn, eventTime}, "_")
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

func AddCorePMMapping(esConf config.ElasticsearchConf) {
	//check if core-pm index exists
	url := esConf.URL + "/_cat/indices/core-pm"
	_, err := httpCall(http.MethodGet, url, esConf.User, esConf.Password, "", nil, defaultTimeout)
	if err != nil && !strings.Contains(err.Error(), "status: 404 Not Found") {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to query elasticsearch")
		return
	}
	if err != nil && strings.Contains(err.Error(), "status: 404 Not Found") {
		err = createMapping("core-pm", esConf)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Unable to cretae mapping for core-pm index")
		}
		return
	}

	log.Info("core-pm index exists, checking mapping")
	//check if mapping is correct
	resp, err := getCorePMMapping("core-pm", "pm_data.*", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to get mapping for core-pm index")
		return
	}

	if !strings.Contains(string(resp), "long") {
		log.Info("Found correct mapping for core-pm")
		return
	}

	log.Info("found incorrect mapping for core-pm, reindexing data...")
	// create mapping on core-pm-temp
	err = createMapping("core-pm-temp", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to creat mapping for core-pm-temp index")
		return
	}
	// reindex core-pm to core-pm-temp
	startTime := time.Now()
	err = reindexData("core-pm", "core-pm-temp", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to reindex data from core-pm to core-pm-temp index")
		return
	}
	endTime := time.Now()
	log.Info("core-pm to core-pm-temp reindex time: ", endTime.Sub(startTime))
	// delete core-pm index
	deleteIndices([]string{"core-pm"}, esConf)
	// create mapping on core-pm
	err = createMapping("core-pm", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to creat mapping for core-pm index")
		return
	}
	// reindex core-pm-temp to core-pm
	startTime = time.Now()
	err = reindexData("core-pm-temp", "core-pm", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to reindex data from core-pm-temp to core-pm index")
		return
	}
	endTime = time.Now()
	log.Info("core-pm-temp to core-pm reindex time: ", endTime.Sub(startTime))
	// delete core-pm-temp index
	deleteIndices([]string{"core-pm-temp"}, esConf)
}

func getCorePMMapping(index, field string, esConf config.ElasticsearchConf) ([]byte, error) {
	//get core-pm mapping
	url := esConf.URL + "/" + index + "/_mapping/field/" + field
	resp, err := httpCall(http.MethodGet, url, esConf.User, esConf.Password, "", nil, defaultTimeout)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to query elasticsearch")
		return nil, err
	}
	return resp, nil
}

func reindexData(source, dest string, esConf config.ElasticsearchConf) error {
	log.Infof("reindexing data for source: %s, destination: %s", source, dest)
	//create mapping on core-pm
	url := esConf.URL + "/_reindex"
	postData := strings.Replace(reindexPostData, "SOURCE", source, -1)
	postData = strings.Replace(postData, "DEST", dest, -1)
	_, err := httpCall(http.MethodPost, url, esConf.User, esConf.Password, postData, nil, 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to query elasticsearch")
		return err
	}
	return nil
}

func createMapping(index string, esConf config.ElasticsearchConf) error {
	log.Infof("creating mapping for %s", index)
	//create mapping on core-pm
	url := esConf.URL + "/" + index
	_, err := httpCall(http.MethodPut, url, esConf.User, esConf.Password, corePMMapping, nil, defaultTimeout)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to query elasticsearch")
		return err
	}
	return nil
}
