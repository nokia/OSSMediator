/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	elasticsearchURL = "http://127.0.0.1:9299"
	testPMData       = `[
    {
      "pm_data": {
        "Cat_M_Accessibility_M8100C0": 0,
        "Cat_M_Accessibility_M8100C10": 0,
        "Cat_M_Accessibility_M8100C11": 0
      },
      "pm_data_source": {
        "edge_id": "test_edge",
        "hw_id": "EB34567",
        "hw_alias": "EB34567",
        "serial_no": "34567",
        "nhg_id": "test_nhg_1",
        "nhg_alias": "test nhg",
        "dn": "NE-MRBTS-111/NE-LNBTS-222/LNCEL-0",
        "timestamp": "2020-11-10T18:30:00Z",
        "technology": "4G"
      }
    },
    {
      "pm_data": {
        "Cat_M_Accessibility_M8100C0": 0,
        "Cat_M_Accessibility_M8100C10": 0,
        "Cat_M_Accessibility_M8100C11": 0
      },
      "pm_data_source": {
        "edge_id": "test_edge2",
        "hw_id": "EB12345",
        "hw_alias": "EB12345",
        "serial_no": "12345",
        "nhg_id": "test_nhg_2",
        "nhg_alias": "test nhg2",
        "dn": "MRBTS-222/NE-LNBTS-4444/LNCEL-0",
        "timestamp": "2020-11-10T18:30:00Z",
        "technology": "4G"
      }
    }
  ]`

	testFMData = `[
    {
      "fm_data": {
        "alarm_identifier": "11111",
        "severity": "minor",
        "specific_problem": "2222",
        "alarm_text": "Failure in connection",
        "additional_text": "TraceConnectionFaultyAl",
        "alarm_state": "CLEARED",
        "event_type": "equipment",
        "event_time": "2020-11-02T06:14:09Z",
        "last_updated_time": "2020-11-02T06:14:10Z",
        "notification_type": "alarmNew"
      },
      "fm_data_source": {
        "edge_id": "98765",
        "hw_id": "EB09856",
        "hw_alias": "eNB7777",
        "serial_no": "45678",
        "nhg_id": "test_nhg_2",
        "nhg_alias": "test nhg2",
        "dn": "MRBTS-456/LNBTS-765",
        "technology": "4G"
      }
    },
    {
      "fm_data": {
        "alarm_identifier": "3333",
        "severity": "minor",
        "specific_problem": "4444",
        "alarm_state": "CLEARED",
        "event_type": "equipment",
        "event_time": "2020-11-02T06:14:10Z",
        "last_updated_time": "2020-11-02T06:14:10Z",
        "notification_type": "alarmClear"
      },
      "fm_data_source": {
        "edge_id": "test_edge",
        "hw_id": "EB1111111",
        "hw_alias": "test eNB",
        "serial_no": "123456",
        "nhg_id": "test_nhg_1",
        "nhg_alias": "test nhg",
        "dn": "MRBTS-123/LNBTS-321",
        "technology": "4G"
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

	fileName := "./pmdata_data.json"
	err := createTestData(fileName, testPMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./pmdata_data.json", elasticsearchURL)
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
	fileName := "./fmdata_RADIO_HISTORY_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./fmdata_RADIO_HISTORY_data.json", elasticsearchURL)
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
		elasticsearchURL = "http://127.0.0.1:9299"
	}()
	fileName := "./fmdata_RADIO_HISTORY_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./fmdata_RADIO_HISTORY_data.json", elasticsearchURL)
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
	fileName := "./fmdata_RADIO_HISTORY_data.json"
	retryTicker = time.NewTicker(2 * time.Millisecond)
	PushFailedDataToElasticsearch(elasticsearchURL)
	time.Sleep(1 * time.Second)
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}

func TestDeleteDataFormElasticsearch(t *testing.T) {
	currTime := currentTime()
	oldCurrentTime := currentTime
	defer func() { currentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(currTime.Year(), currTime.Month(), currTime.Day()+1, 0, 59, 59, 0, time.Local)
	}
	currentTime = myCurrentTime
	go DeleteDataFormElasticsearch(elasticsearchURL, 0)
	time.Sleep(5 * time.Second)

	searchURL := elasticsearchURL + "/" + strings.Join(indexList, ",") + "/_search"
	request, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		t.Error(err)
	}

	response, err := newNetClient().Do(request)
	if err != nil {
		t.Error(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Received response' status code: %d, status: %s", response.StatusCode, response.Status)
	}

	bodyBytes, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}
	bodyString := string(bodyBytes)
	t.Log(bodyString)
	if !strings.Contains(string(bodyBytes), "\"hits\":{\"total\":{\"value\":0,\"relation\":\"eq\"}") {
		t.Fail()
	}
}
