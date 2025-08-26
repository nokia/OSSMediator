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
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type pmResponse struct {
	PMData       map[string]interface{} `json:"pm_data"`
	PMDataSource map[string]interface{} `json:"pm_data_source"`
	Timestamp    time.Time              `json:"timestamp"`
}

type fmResponse struct {
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
	fileName := path.Base(filePath)
	baseMetricType := strings.Split(fileName, "_")[1]
	baseMetricType = strings.ToLower(baseMetricType)

	file, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading file: %s", filePath)
		return
	}
	defer file.Close()

	index := strings.Join([]string{baseMetricType, "fm"}, "-")
	var postData string
	dec := json.NewDecoder(file)

	// read open bracket
	t, err := dec.Token()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while getting json token: %s", filePath)
		return
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		log.WithFields(log.Fields{"error": err}).Errorf("Invalid file %s, array starting not found", filePath)
		return
	}

	// while the array contains values
	i := 0
	for dec.More() {
		var resp fmResponse
		var id, metricType string
		err = dec.Decode(&resp)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Error while decoding json: %s", filePath)
			return
		}
		i++

		if resp.FMData["event_time"] == nil {
			continue
		}
		dn := resp.FMDataSource["dn"].(string)
		eventTime := resp.FMData["event_time"].(string)
		hwID := resp.FMDataSource["hw_id"].(string)
		alarmID := resp.FMData["alarm_identifier"].(string)
		if baseMetricType == "all" {
			metricType = resp.FMDataSource["metric_type"].(string)
			metricType = strings.ToLower(metricType)
			index = strings.Join([]string{metricType, "fm"}, "-")
		}

		if baseMetricType == "radio" || metricType == "radio" {
			specificProb := resp.FMData["specific_problem"].(string)
			id = strings.Join([]string{hwID, dn, alarmID, specificProb, eventTime}, "_")
		} else {
			faultID := resp.FMData["fault_id"].(string)
			keys := []string{metricType, hwID, dn, alarmID}
			if faultID != "" {
				keys = append(keys, faultID)
			}
			keys = append(keys, eventTime)
			id = strings.Join(keys, "_")
		}
		resp.Timestamp = time.Now()
		source, _ := json.Marshal(resp)
		postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
		postData += string(source) + "\n"
		if i%elkNoOfRecordsPerAPI == 0 {
			pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
			postData = ""
		}
	}

	if postData != "" {
		pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
		postData = ""
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while getting json token: %s", filePath)
		return
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		log.WithFields(log.Fields{"error": err}).Errorf("Invalid file %s, array ending not found", filePath)
		return
	}
}

func pushPMData(filePath string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + elkBulkAPI
	log.Infof("Pushing data from %s to elasticsearch", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading file: %s", filePath)
		return
	}
	defer file.Close()

	fileName := path.Base(filePath)
	metricType := strings.Split(fileName, "_")[1]
	metricType = strings.ToLower(metricType)
	var postData string
	currTime := time.Now().UTC()
	dec := json.NewDecoder(file)

	// read open bracket
	t, err := dec.Token()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while getting json token: %s", filePath)
		return
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		log.WithFields(log.Fields{"error": err}).Errorf("Invalid file %s, array starting not found", filePath)
		return
	}

	// while the array contains values
	i := 0
	for dec.More() {
		var resp pmResponse
		err = dec.Decode(&resp)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Error while decoding json: %s", filePath)
			return
		}
		i++

		hwID := resp.PMDataSource["hw_id"].(string)
		dn := resp.PMDataSource["dn"].(string)
		eventTime := resp.PMDataSource["timestamp"].(string)
		technology := resp.PMDataSource["technology"].(string)
		var objectType string
		for k := range resp.PMData {
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

		resp.Timestamp = time.Now()
		source, _ := json.Marshal(resp)
		postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
		postData += string(source) + "\n"
		if i%elkNoOfRecordsPerAPI == 0 {
			pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
			postData = ""
		}
	}

	if postData != "" {
		pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
		postData = ""
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while getting json token: %s", filePath)
		return
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		log.WithFields(log.Fields{"error": err}).Errorf("Invalid file %s, array ending not found", filePath)
		return
	}
}

func AddPMMapping(esConf config.ElasticsearchConf, index string) {
	//check if core-pm index exists
	url := esConf.URL + "/_cat/indices/" + index
	_, err := httpCall(http.MethodGet, url, esConf.User, esConf.Password, nil, nil, defaultTimeout)
	if err != nil && !strings.Contains(err.Error(), "status: 404 Not Found") {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to query elasticsearch")
		return
	}
	if err != nil && strings.Contains(err.Error(), "status: 404 Not Found") {
		err = createMapping(index, esConf)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Unable to cretae mapping for %s index ", index)
		}
		return
	}

	log.Info(index + " index exists, checking mapping")
	//check if mapping is correct
	resp, err := getPMMapping(index, "pm_data.*", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to get mapping for %s index", index)
		return
	}

	if !strings.Contains(string(resp), "long") {
		log.Info("Found correct mapping for : " + index)
		return
	}

	log.Info("found incorrect mapping for " + index + ", reindexing data...")
	// create mapping on core-pm-temp
	err = createMapping(index+"-temp", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to creat mapping for %s-temp index", index)
		return
	}
	// reindex core-pm to core-pm-temp
	startTime := time.Now()
	err = reindexData(index, index+"-temp", esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to reindex data from %s to %s-temp index", index, index)
		return
	}
	endTime := time.Now()
	log.Info(index+" to "+index+"-temp reindex time: ", index, index, endTime.Sub(startTime))
	// delete core-pm index
	deleteIndices([]string{index}, esConf)
	// create mapping on core-pm
	err = createMapping(index, esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to create mapping for %s", index)
		return
	}
	// reindex core-pm-temp to core-pm
	startTime = time.Now()
	err = reindexData(index+"-temp", index, esConf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to reindex data from %s-temp to %s index", index, index)
		return
	}
	endTime = time.Now()
	log.Info(index+"-temp to "+index+" reindex time: ", endTime.Sub(startTime))
	// delete -pm-temp index
	deleteIndices([]string{index + "-temp"}, esConf)
}

func getPMMapping(index, field string, esConf config.ElasticsearchConf) ([]byte, error) {
	//get core-pm mapping
	url := esConf.URL + "/" + index + "/_mapping/field/" + field
	resp, err := httpCall(http.MethodGet, url, esConf.User, esConf.Password, nil, nil, defaultTimeout)
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
	_, err := httpCall(http.MethodPost, url, esConf.User, esConf.Password, &postData, nil, 0)
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
	mapping := corePMMapping
	_, err := httpCall(http.MethodPut, url, esConf.User, esConf.Password, &mapping, nil, defaultTimeout)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to query elasticsearch")
		return err
	}
	return nil
}
