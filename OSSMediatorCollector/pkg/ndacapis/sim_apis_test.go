/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package ndacapis

import (
	"bytes"
	"collector/pkg/config"
	"collector/pkg/utils"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	simAPIResponse = `{
  "status": {
    "status_code": "SUCCESS",
    "status_description": {
      "description_code": "NOT_SPECIFIED",
      "description": "Successful fetch of subscriptions"
    }
  },
  "email_id": "user@nokia.com",
  "page_response": {
    "page_details": {
      "page_number": 1,
      "page_size": 1
    },
    "total_pages": 1
  },
  "subsc": [
    {
      "imsi": "12345",
      "imsi_alias": "ue",
      "icc_id": "333333",
      "imei": "222222",
      "apn_id": "DEFAULTAPN",
      "apn_name": "internet",
      "qos_group_id": "DEFAULTQOS",
      "qos_profile_name": "internet_DEFAULT",
      "private_network_details": [
        {
          "nhg_id": "test_nhg",
          "status": "ACTIVE",
          "status_description": "string",
          "ue_status": "UE_CONNECTED",
          "ue_report_timestamp": "2020-02-25T10:48:34.859Z",
          "nhg_alias": "Network1",
          "feature_status": {
            "imsi": "123456",
            "static_ip_status": "IN_SYNC",
            "qos_status": "IN_SYNC",
            "imei_status": "OUT_OF_SYNC",
            "multiple_apn_status": "IN_SYNC"
          }
        }
      ]
    }
  ],
  "requested_sims": 1
}`

	apSimsAPIResponse = `{
  "status": {
    "status_code": "SUCCESS",
    "status_description": {
      "description_code": "NOT_SPECIFIED",
      "description": "Successful fetch of subscriptions"
    }
  },
  "access_point_imsi_details": [
    {
      "access_point_hw_id": "test-hw1",
      "imsi_details": [
        {
          "imsi": "111111",
          "icc_id": "12345",
          "ue_report_timestamp": "2020-02-25T10:48:34.859Z",
          "imsi_alias": "test1"
        },
        {
          "imsi": "222222",
          "icc_id": "987654",
          "ue_report_timestamp": "2020-02-25T11:48:34.859Z",
          "imsi_alias": "test2"
        }
      ]
    }
  ]
}`
)

func TestCallSimAPI(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken: "accessToken",
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, simAPIResponse)
	}))
	defer testServer.Close()

	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	apiConf := config.APIConf{API: "/sims", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	defer os.RemoveAll(user.ResponseDest)
	CreateHTTPClient("", true)

	oldCurrentTime := utils.CurrentTime
	defer func() { utils.CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	utils.CurrentTime = myCurrentTime
	fetchSimData(&apiConf, &user, 123)

	fileName := "./tmp/sims/sims_testuser@nokia.com_response_" + strconv.Itoa(int(utils.CurrentTime().Unix())) + ".json"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Error("File not found,  ", err)
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error(err)
	}
	if len(data) == 0 {
		t.Error("Found empty file")
	}
}

func TestCallSimAPIWithWrongURL(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken: "accessToken",
	}
	m := map[string]config.OrgAccDetails{}
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails.OrgUUID = "org_uuid_1"
	orgAcc.OrgDetails.OrgAlias = "org_alias_1"
	orgAcc.AccDetails.AccUUID = "acc_uuid_1"
	orgAcc.AccDetails.AccAlias = "acc_alias_1"

	m["test_nhg1"] = orgAcc
	user.NhgIDsABAC = m

	config.Conf = config.Config{
		BaseURL: "http://localhost",
	}
	apiConf := config.APIConf{API: "/sims/{nhg_id}", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	CreateHTTPClient("", true)

	oldCurrentTime := utils.CurrentTime
	defer func() { utils.CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	utils.CurrentTime = myCurrentTime
	fetchSimData(&apiConf, &user, 123)
	fileName := "./tmp/sims/sims_testuser@nokia.com_response_" + strconv.Itoa(int(utils.CurrentTime().Unix())) + ".json"
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Fail()
	}
}

func TestGetSimsDataWithInactiveSession(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: false}
	fetchSimData(&config.APIConf{API: "/sims/{nhg_id}", Interval: 15}, &user, 1234)
	if !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestCallAPSimAPIWithWrongURL(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken: "accessToken",
	}
	m := map[string]config.OrgAccDetails{}
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails.OrgUUID = "org_uuid_1"
	orgAcc.OrgDetails.OrgAlias = "org_alias_1"
	orgAcc.AccDetails.AccUUID = "acc_uuid_1"
	orgAcc.AccDetails.AccAlias = "acc_alias_1"

	m["test_nhg1"] = orgAcc
	user.NhgIDsABAC = m

	config.Conf = config.Config{
		BaseURL: "http://localhost",
	}
	apiConf := config.APIConf{API: "/access-point-sims", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	CreateHTTPClient("", true)

	oldCurrentTime := utils.CurrentTime
	defer func() { utils.CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	utils.CurrentTime = myCurrentTime
	fetchSimData(&apiConf, &user, 123)
	fileName := "./tmp/access-point-sims/access-point-sims_testuser@nokia.com_response_" + strconv.Itoa(int(utils.CurrentTime().Unix())) + ".json"
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Fail()
	}
}

func TestGetAPSimsDataWithInactiveSession(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: false}
	fetchSimData(&config.APIConf{API: "/access-point-sims", Interval: 15}, &user, 1234)
	if !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestCallAPSimAPI(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken: "accessToken",
	}
	m := map[string]config.OrgAccDetails{}
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails.OrgUUID = "org_uuid_1"
	orgAcc.OrgDetails.OrgAlias = "org_alias_1"
	orgAcc.AccDetails.AccUUID = "acc_uuid_1"
	orgAcc.AccDetails.AccAlias = "acc_alias_1"

	m["test-hw1"] = orgAcc
	user.HwIDsABAC = m

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, apSimsAPIResponse)
	}))
	defer testServer.Close()

	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	apiConf := config.APIConf{API: "/access-point-sims", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	defer os.RemoveAll(user.ResponseDest)
	CreateHTTPClient("", true)

	oldCurrentTime := utils.CurrentTime
	defer func() { utils.CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	utils.CurrentTime = myCurrentTime
	fetchSimData(&apiConf, &user, 123)

	fileName := "./tmp/access-point-sims/access-point-sims_testuser@nokia.com_response_" + strconv.Itoa(int(utils.CurrentTime().Unix())) + ".json"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Error("File not found,  ", err)
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error(err)
	}
	if len(data) == 0 {
		t.Error("Found empty file")
	}
}
