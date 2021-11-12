/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"bytes"
	"elasticsearchplugin/pkg/config"
	"net/http"
	"os"
	"strconv"
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
        "notification_type": "alarmNew",
		"fault_id": "1"
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
        "notification_type": "alarmClear",
		"fault_id": ""
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

	testNhgData = `[
    {
      "nhg_id": "123456",
      "nhg_alias": "test_nhg",
      "deployment_type": "SMALL",
      "deploy_readiness_status": 1,
      "clusters": [
        {
          "cluster_id": "1",
          "cluster_alias": "cluster1",
          "hw_set": [
            {
              "hw_id": "AA-BB-CC-DD-EE-FF::UpBoard::UpBoard-1",
              "hw_type": "EdgeCloud",
              "serial_no": "AA-BB-CC-DD-EE-FF"
            }
          ],
          "cluster_status": "ACTIVE",
          "latitude": 45,
          "longitude": 45,
          "nw_config_status": "ACTIVE",
          "edge_connection_status": "EDGE_CONN_STATUS_UNAVAILABLE",
          "cluster_slice_status_modify_time": "2021-02-20T07:36:07.976Z",
          "slice_id": "1222-13"
        }
      ],
      "nhg_config_status": "NW_CFG_UNAVAILABLE",
      "nhg_slice_status_modify_time": "2021-02-20T07:36:07.976Z",
      "network_type": "NETWORK_TYPE_4G",
      "usb_install_status": "USB_STATUS_UNAVAILABLE",
      "sla": "SLA_BASIC"
    }
  ]`

	testSimData = `{
  "email_id": "testuser@nokia.com",
  "subsc": [
    {
      "imsi": "12345",
      "imsi_alias": "imsi",
      "icc_id": "12345",
      "imei": "12345",
      "apn_id": "DEFAULTAPN",
      "apn_name": "internet",
      "qos_group_id": "DEFAULTQOS",
      "qos_profile_name": "internet_DEFAULT",
      "private_network_details": [
        {
          "nhg_id": "123456",
          "status": "ACTIVE",
          "status_description": "string",
          "ue_status": "UE_CONNECTED",
          "ue_report_timestamp": "2021-02-25T10:48:34.859Z",
          "nhg_alias": "Network1",
          "feature_status": {
            "imsi": "1111111111",
            "static_ip_status": "IN_SYNC",
            "qos_status": "IN_SYNC",
            "imei_status": "IN_SYNC",
            "multiple_apn_status": "IN_SYNC"
          }
        }
      ]
    }
  ],
  "requested_sims": 1
}`

	testApSimsData = `[
  {
    "access_point_hw_id": "test-hw",
    "imsi_details": [
      {
        "icc_id": "12345",
        "imsi": "999999",
        "imsi_alias": "Test",
        "ue_report_timestamp": "2021-09-09T15:25:19.376Z"
      },
      {
        "icc_id": "987654",
        "imsi": "888888",
        "imsi_alias": "Test 2",
        "ue_report_timestamp": "2021-09-15T10:52:19.085Z"
      }
    ]
  }
]`
)

func TestPushPMDataToElasticsearch(t *testing.T) {
	fileName := "./pmdata_data.json"
	err := createTestData(fileName, testPMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: elasticsearchURL,
	}

	PushData(fileName, esConf)
	time.Sleep(2 * time.Second)
	searchResult, err := searchOnElastic([]string{"4g-pm*"})
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":2,\"relation\":\"eq\"}") {
		t.Fail()
	}
}

func TestPushRadioFMDataToElasticsearch(t *testing.T) {
	fileName := "./fmdata_RADIO_HISTORY_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: elasticsearchURL,
	}

	PushData(fileName, esConf)
	time.Sleep(2 * time.Second)
	searchResult, err := searchOnElastic([]string{"radio-fm*"})
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":2,\"relation\":\"eq\"}") {
		t.Fail()
	}
}

func TestPushDACFMDataToElasticsearch(t *testing.T) {
	fileName := "./fmdata_DAC_HISTORY_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: elasticsearchURL,
	}

	PushData(fileName, esConf)
	time.Sleep(2 * time.Second)
	searchResult, err := searchOnElastic([]string{"dac-fm*"})
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":2,\"relation\":\"eq\"}") {
		t.Fail()
	}
}

func TestPushCoreFMDataToElasticsearch(t *testing.T) {
	fileName := "./fmdata_CORE_HISTORY_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: elasticsearchURL,
	}

	PushData(fileName, esConf)
	time.Sleep(2 * time.Second)
	searchResult, err := searchOnElastic([]string{"core-fm*"})
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":2,\"relation\":\"eq\"}") {
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
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: "http://localhost:12345",
	}

	PushData(fileName, esConf)
	if !strings.Contains(buf.String(), "Unable to push data to elasticsearch, will be retried later") {
		t.Fail()
	}

	retryTicker = time.NewTicker(2 * time.Millisecond)
	esConf.URL = "http://127.0.0.1:9299"
	PushFailedData(esConf)
	time.Sleep(1 * time.Second)
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}

func searchOnElastic(indices []string) (string, error) {
	searchURL := elasticsearchURL + "/" + strings.Join(indices, ",") + "/_search"
	resp, err := httpCall(http.MethodGet, searchURL, "", "", "", nil)
	if err != nil {
		return "", err
	}
	searchResult := string(resp)
	return searchResult, nil
}

func createTestData(filename, content string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	if err != nil {
		return err
	}
	return nil
}

func TestPushSimsDataToElasticsearch(t *testing.T) {
	fileName := "./sims_12345_data.json"
	err := createTestData(fileName, testSimData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: elasticsearchURL,
	}

	PushData(fileName, esConf)
	time.Sleep(2 * time.Second)
	searchResult, err := searchOnElastic([]string{"sims-data"})
	t.Log(searchResult)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":1,\"relation\":\"eq\"}") {
		t.Fail()
	}
}

func TestPushAPSimsDataToElasticsearch(t *testing.T) {
	fileName := "./access-point-sims_12345_data.json"
	err := createTestData(fileName, testApSimsData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: elasticsearchURL,
	}

	PushData(fileName, esConf)
	time.Sleep(2 * time.Second)
	searchResult, err := searchOnElastic([]string{"ap-sims-data"})
	t.Log(searchResult)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":2,\"relation\":\"eq\"}") {
		t.Fail()
	}
}

func TestPushNhgDataToElasticsearch(t *testing.T) {
	fileName := "./network-hardware-groups_testuser@nokia.com_data.json"
	err := createTestData(fileName, testNhgData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	esConf := config.ElasticsearchConf{
		URL: elasticsearchURL,
	}

	PushData(fileName, esConf)
	time.Sleep(2 * time.Second)
	searchResult, err := searchOnElastic([]string{"nhg-data"})
	t.Log(searchResult)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":1,\"relation\":\"eq\"}") {
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
	esConf := config.ElasticsearchConf{
		URL:                   elasticsearchURL,
		DataRetentionDuration: 0,
	}
	deletionTime := currentTime().AddDate(0, 0, -1*esConf.DataRetentionDuration).UTC().Format(timestampFormat)
	deleteData(indexList, deletionTime, esConf)
	time.Sleep(2 * time.Second)

	searchResult, err := searchOnElastic(indexList)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":0,\"relation\":\"eq\"}") {
		t.Fail()
	}
}

func TestDeleteOldIndicesFormElasticsearch(t *testing.T) {
	esConf := config.ElasticsearchConf{
		URL:                   elasticsearchURL,
		DataRetentionDuration: 60,
	}

	delTime := time.Now().AddDate(0, 0, -1*esConf.DataRetentionDuration)
	indexSuffix := strconv.Itoa(int(delTime.Month())) +"-"+ strconv.Itoa(delTime.Year())
	testIndices := []string{"4g-pm-"+indexSuffix, "5g-pm-"+indexSuffix}
	for _, index := range testIndices {
		indexCreateURL := elasticsearchURL + "/" + index
		_, err := httpCall(http.MethodPut, indexCreateURL, "", "", "", nil)
		if err != nil {
			t.Error(err)
		}
	}

	indices := getOldIndices(esConf)
	if len(indices) == 0 {
		t.Error("found no indices")
	}
	deleteIndices(indices, esConf)
	time.Sleep(2 * time.Second)

	searchResult, err := searchOnElastic(indices)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(searchResult, "\"hits\":{\"total\":{\"value\":0,\"relation\":\"eq\"}") {
		t.Fail()
	}
}

func TestGetNextTickDuration(t *testing.T) {
	currTime := currentTime()
	oldCurrentTime := currentTime
	defer func() { currentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(currTime.Year(), currTime.Month(), currTime.Day()+1, 0, 0, 0, 0, time.Local)
	}
	currentTime = myCurrentTime
	nextTick := getNextTickDuration()
	if nextTick != time.Hour {
		t.Fail()
	}
}
