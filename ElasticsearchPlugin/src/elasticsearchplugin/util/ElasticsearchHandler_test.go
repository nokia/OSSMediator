/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	elasticsearchURL = "http://127.0.0.1:9200"
	testPMData       = `[
	{
	  "_index": "test_index",
	  "_id": "12345",
	  "_source": {
		"Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
		"edgename": "test edge",
		"SIG_SctpCongestionDuration": 0,
		"neserialno": "4355",
		"nename": "test",
		"edge_id": "edge_id",
		"SIG_SctpUnavailableDuration": 0,
		"edge_hostname": "localhost",
		"timestamp": "2018-11-22T16:15:00+05:30"
	  }
	},
	{
	  "_index": "test_index",
	  "_id": "124576",
	  "_source": {
		"Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
		"edgename": "test edge 2",
		"EQPT_MaxMeLoad": 0.5,
		"neserialno": "3456",
		"EQPT_MeanCpuUsage_0": 5,
		"EQPT_MeanCpuUsage_1": 5,
		"EQPT_MaxCpuUsage_1": 99,
		"edge_id": "edge_id2",
		"EQPT_MeanMeLoad": 0.4,
		"edge_hostname": "localhost",
		"ObjectType": "ManagedElement",
		"UserLabel": "temp",
		"nename": "476486",
		"EQPT_MaxCpuUsage_0": 100,
		"EQPT_MaxMemUsage": 59,
		"EQPT_MeanMemUsage": 59,
		"timestamp": "2018-11-22T16:15:00+05:30"
	  }
	}
  ]`

	testFMData = `[  
    {
       "_index":"test_index",
       "_type":"",
       "_id":"12345",
       "_source":{
		  "Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
          "EventType":"Processing Error Alarm",
          "AlarmState":0,
          "SpecificProblem":"Time synchronization failed",
          "LastUpdatedTime":"2018-09-23T16:04:06Z",
          "edge_id":"edge_id",
          "Severity":"Minor",
          "NESerialNo":"98675",
          "NeName":"",
          "AdditionalText":"Unspecified",
          "edge_hostname":"localhost",
          "ProbableCause":"Time synchronization failed",
          "AlarmText":"Time synchronization failed",
          "EventTime":"2018-09-23T21:33:35+05:30",
          "NotificationType":"NewAlarm",
          "AlarmIdentifier":"11191",
          "EdgeName":""
       }
    },
    {
       "_index":"test_index",
       "_type":"",
       "_id":"12345",
       "_source":{
          "Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
          "EventType":"Processing Error Alarm",
          "AlarmState":0,
          "SpecificProblem":"Time synchronization failed",
          "LastUpdatedTime":"2018-09-24T10:58:06Z",
          "edge_id":"edge_id",
          "Severity":"Minor",
          "NESerialNo":"564357",
          "NeName":"",
          "AdditionalText":"Unspecified",
          "edge_hostname":"localhost",
          "ProbableCause":"Time synchronization failed",
          "AlarmText":"Time synchronization failed",
          "EventTime":"2018-09-24T16:28:01+05:30",
          "NotificationType":"ClearedAlarm",
          "AlarmIdentifier":"11191",
          "EdgeName":""
       }
    }
 ]`
)

func TestPushPMDataToElasticsearch(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	fileName := "./pm_data.json"
	err := createTestData(fileName, testPMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./pm_data.json", elasticsearchURL)
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}

func TestPushFMDataToElasticsearch(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	fileName := "./fm_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./fm_data.json", elasticsearchURL)
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}

func TestPushFMDataToInvalidElasticsearch(t *testing.T) {
	elasticsearchURL = "http://localhost:12345"
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
		elasticsearchURL = "http://127.0.0.1:9200"
	}()
	fileName := "./fm_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./fm_data.json", elasticsearchURL)
	if !strings.Contains(buf.String(), "Unable to push data to elasticsearch, will be retried later") {
		t.Fail()
	}
}

func TestPushFailedDataToElasticsearch(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	fileName := "./fm_data.json"
	retryTicker = time.NewTicker(2 * time.Millisecond)
	PushFailedDataToElasticsearch(elasticsearchURL)
	time.Sleep(1 * time.Second)
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}
