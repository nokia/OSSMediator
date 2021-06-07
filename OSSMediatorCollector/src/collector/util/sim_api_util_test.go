package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
)

func TestCallSimAPI(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true, ResponseDest: "./tmp"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, simAPIResponse)
	}))
	defer testServer.Close()

	apiConf := APIConf{API: "/sims", Interval: 15}
	CreateResponseDirectory(user.ResponseDest, apiConf.API)
	CreateHTTPClient("", true)

	oldCurrentTime := currentTime
	defer func() { currentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	currentTime = myCurrentTime
	fetchSimData(testServer.URL, &apiConf, &user, 123)

	fileName := "./tmp/sims/sims_testuser@nokia.com_response_"+ strconv.Itoa(int(currentTime().Unix())) +".json"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Error("File not found,  ", err)
	}
	defer os.Remove(fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error(err)
	}
	if len(data) == 0 {
		t.Error("Found empty file")
	}
}

func TestCallSimAPIWithWrongURL(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true, ResponseDest: "./tmp"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
	}
	user.nhgIDs = []string{"test_nhg1"}

	apiConf := APIConf{API: "/sims/{nhg_id}", Interval: 15}
	CreateResponseDirectory(user.ResponseDest, apiConf.API)
	CreateHTTPClient("", true)

	oldCurrentTime := currentTime
	defer func() { currentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	currentTime = myCurrentTime
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	fetchSimData("http://localhost", &apiConf, &user, 123)
	t.Log(buf.String())
	fileName := "./tmp/sims/sims_testuser@nokia.com_response_"+ strconv.Itoa(int(currentTime().Unix())) +".json"
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Fail()
	}
}
