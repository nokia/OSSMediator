/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	response = `{
		"data": "[{\"_index\":\"test-index\",\"_id\":\"12345\",\"_source\":{\"EventType\":\"Equipment Alarm\",\"LastUpdatedTime\":\"2018-12-12T03:31:46Z\",\"Severity\":\"Minor\",\"EventTime\":\"2018-12-11T23:15:13+05:30:00\",\"NotificationType\":\"NewAlarm\"}},{\"_index\":\"test-index\",\"_id\":\"12345\",\"_source\":{\"EventType\":\"communications\",\"LastUpdatedTime\":\"2018-12-12T08:57:43+05:30\",\"Severity\":\"major\",\"EventTime\":\"2018-12-11T12:52:34+05:30\",\"NotificationType\":\"alarmNew\"}}]",
		"next_record": 0,
		"num_of_records": 2,
		"status": {
		  "status_code": "SUCCESS",
		  "status_description": {
			"description": "Records transferred successfully",
			"description_code": "NOT_SPECIFIED"
		  }
		},
		"total_num_records": 2,
		"type": "fmdata"
	  }`

	response1 = `{
		"data": "[{\"_index\":\"test-index\",\"_id\":\"12345\",\"_source\":{\"EventType\":\"Equipment Alarm\",\"LastUpdatedTime\":\"2018-12-12T03:31:46Z\",\"Severity\":\"Minor\",\"EventTime\":\"2018-12-11T23:15:13+05:30:00\",\"NotificationType\":\"NewAlarm\"}},{\"_index\":\"test-index\",\"_id\":\"12345\",\"_source\":{\"EventType\":\"communications\",\"LastUpdatedTime\":\"2018-12-12T08:57:43+05:30\",\"Severity\":\"major\",\"EventTime\":\"2018-12-11T12:52:34+05:30\",\"NotificationType\":\"alarmNew\"}}]",
		"next_record": 2,
		"num_of_records": 2,
		"status": {
		  "status_code": "SUCCESS",
		  "status_description": {
			"description": "Records transferred successfully",
			"description_code": "NOT_SPECIFIED"
		  }
		},
		"total_num_records": 4,
		"type": "fmdata"
	  }`

	response2 = `{
		"data": "[{\"_index\":\"test-index\",\"_id\":\"12345\",\"_source\":{\"EventType\":\"Equipment Alarm\",\"LastUpdatedTime\":\"2018-12-12T03:31:46Z\",\"Severity\":\"Minor\",\"EventTime\":\"2018-12-11T23:15:13+05:30:00\",\"NotificationType\":\"NewAlarm\"}},{\"_index\":\"test-index\",\"_id\":\"12345\",\"_source\":{\"EventType\":\"communications\",\"LastUpdatedTime\":\"2018-12-12T08:57:43+05:30\",\"Severity\":\"major\",\"EventTime\":\"2018-12-11T12:52:34+05:30\",\"NotificationType\":\"alarmNew\"}}]",
		"next_record": 0,
		"num_of_records": 2,
		"status": {
		  "status_code": "SUCCESS",
		  "status_description": {
			"description": "Records transferred successfully",
			"description_code": "NOT_SPECIFIED"
		  }
		},
		"total_num_records": 4,
		"type": "fmdata"
	  }`

	listNhgResp = `
	{
  		"status": {
    		"status_code": "SUCCESS",
    		"status_description": {
      			"description_code": "NOT_SPECIFIED",
      			"description": "Fetch Successful for the requested user"
    		}
  		},
  		"nhg_info": [
    		{
      			"nhg_id": "test_nhg_1",
      			"nhg_config_status": "NW_CFG_UNAVAILABLE"
    		},
    		{
      			"nhg_id": "test_nhg_2",
      			"nhg_config_status": "ACTIVE"
			},
    		{
      			"nhg_id": "test_nhg_3",
				"nhg_config_status": "CONFIG_SUCCESS"
			},
    		{
      			"nhg_id": "test_nhg_4",
      			"nhg_config_status": "CONFIG_READY"
    		},
    		{
      			"nhg_id": "test_nhg_5",
      			"nhg_config_status": "NW_CFG_UNAVAILABLE"
    		}
  		]
	}`
)

func TestCreateHTTPClientForSkipTLS(t *testing.T) {
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	CreateHTTPClient("", true)
	if !strings.Contains(buf.String(), "Skipping TLS authentication") {
		t.Fail()
	}
}

func TestCreateHTTPClientWithRootCRT(t *testing.T) {
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	CreateHTTPClient("", false)
	if !strings.Contains(buf.String(), "TLS authentication using root certificates") {
		t.Fail()
	}
}

func TestCreateHTTPClientWithNonExistingCRTFile(t *testing.T) {
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	CreateHTTPClient("tmp.crt", false)
	if !(strings.Contains(buf.String(), "Error while reading server certificate file") && strings.Contains(buf.String(), "TLS authentication using root certificates")) {
		t.Fail()
	}
}

func TestCreateHTTPClientWithCRTFile(t *testing.T) {
	tmpfile := createTmpFile(".", "crt", []byte(``))
	defer os.Remove(tmpfile)
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	CreateHTTPClient(tmpfile, false)
	if !strings.Contains(buf.String(), "Using CA certificate "+tmpfile) {
		t.Fail()
	}

}

func TestWriteResponseWithWrongDir(t *testing.T) {
	response := &GetAPIResponse{
		Data:            "test data",
		NextRecord:      0,
		NumOfRecords:    10,
		Type:            pmResponseType,
		TotalNumRecords: 10,
	}
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	writeResponse(response, user, "", 123)
	if !strings.Contains(buf.String(), "File creation failed") {
		t.Fail()
	}
}

func TestWriteResponseForPM(t *testing.T) {
	response := &GetAPIResponse{
		Data:            `{"data":"test data"}`,
		NextRecord:      0,
		NumOfRecords:    10,
		Type:            pmResponseType,
		TotalNumRecords: 10,
	}

	user := &User{Email: "testuser@nokia.com", ResponseDest: "./tmp"}
	CreateResponseDirectory(user.ResponseDest, "/pm")
	writeResponse(response, user, "/pm", 123)
	defer os.RemoveAll(user.ResponseDest)
	files, err := ioutil.ReadDir(user.ResponseDest + "/pm")
	if err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile(user.ResponseDest + "/pm/" + files[0].Name())
	if err != nil || len(data) == 0 {
		t.Fail()
	}
}

func TestWriteResponseForFM(t *testing.T) {
	response := &GetAPIResponse{
		Data:            `{"data":"test data"}`,
		NextRecord:      0,
		NumOfRecords:    10,
		Type:            "fmdata",
		TotalNumRecords: 10,
	}

	user := &User{Email: "testuser@nokia.com", ResponseDest: "./tmp"}
	CreateResponseDirectory(user.ResponseDest, "/fm")
	writeResponse(response, user, "/fm", 123)
	defer os.RemoveAll(user.ResponseDest)
	files, err := ioutil.ReadDir(user.ResponseDest + "/fm")
	if err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile(user.ResponseDest + "/fm/" + files[0].Name())
	if err != nil || len(data) == 0 {
		t.Fail()
	}
}

//Unit test for callAPI and doRequest
func TestCallAPIForInvalidCase(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Token sent is empty. Invalid Token"
				}
			}
		}`)
	}))
	defer testServer.Close()

	CreateHTTPClient("", true)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	response, err := callAPI(testServer.URL, &user, "", startTime, endTime, 0, 100, "", 123)
	if err == nil || response != nil || !strings.Contains(err.Error(), "Error while validating response status") {
		t.Fail()
	}
}

func TestCallAPIForInvalidURL(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}

	CreateHTTPClient("", true)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	response, err := callAPI(":", &user, "", startTime, endTime, 0, 100, "", 123)
	if err == nil || response != nil || !strings.Contains(err.Error(), "missing protocol scheme") {
		t.Fail()
	}
}

// //Unit test for callAPI and doRequest
func TestCallAPI(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, response)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	resp, err := callAPI(testServer.URL, &user, "", startTime, endTime, 0, 100, "", 123)
	if err != nil || resp.Status.StatusCode != "SUCCESS" || resp.Type != "fmdata" || resp.TotalNumRecords != 2 || resp.NumOfRecords != 2 || resp.NextRecord != 0 || len(resp.Data) == 0 {
		t.Fail()
	}
}

func TestCallAPIWithInvalidResponse(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	resp, err := callAPI(testServer.URL, &user, "", startTime, endTime, 0, 100, "", 123)
	if resp != nil && strings.Contains(err.Error(), "Unable to decode response") {
		t.Fail()
	}
}

func TestCallAPIWIthLastReceivedFile(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, response)
	}))

	tmpFile := lastReceivedDataTime + "_" + user.Email + "_" + path.Base(testServer.URL)
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	_, err = file.Write([]byte("2018-03-14T13:55:00+05:30"))
	if err != nil {
		t.Error(err)
	}
	defer testServer.Close()
	defer os.Remove(tmpFile)
	CreateHTTPClient("", false)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	resp, err := callAPI(testServer.URL, &user, "", startTime, endTime, 0, 100, "", 123)
	if err != nil || resp.Status.StatusCode != "SUCCESS" || resp.Type != "fmdata" || resp.TotalNumRecords != 2 || resp.NumOfRecords != 2 || resp.NextRecord != 0 || len(resp.Data) == 0 {
		t.Fail()
	}
}

func TestCallAPIWithSkipCert(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, response)
	}))
	defer testServer.Close()
	CreateHTTPClient("", true)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	resp, err := callAPI(testServer.URL, &user, "", startTime, endTime, 0, 100, "", 123)
	if err != nil || resp.Status.StatusCode != "SUCCESS" || resp.Type != "fmdata" || resp.TotalNumRecords != 2 || resp.NumOfRecords != 2 || resp.NextRecord != 0 || len(resp.Data) == 0 {
		t.Fail()
	}
}

func TestCallAPIWithCert(t *testing.T) {
	tmpfile := createTmpFile(".", "crt", []byte(``))
	CreateHTTPClient(tmpfile, false)
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, response)
	}))
	defer testServer.Close()
	defer os.Remove(tmpfile)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	resp, err := callAPI(testServer.URL, &user, "", startTime, endTime, 0, 100, "", 123)
	if err != nil || resp.Status.StatusCode != "SUCCESS" || resp.Type != "fmdata" || resp.TotalNumRecords != 2 || resp.NumOfRecords != 2 || resp.NextRecord != 0 || len(resp.Data) == 0 {
		t.Fail()
	}
}

func TestCallAPIWithErrorStatusCode(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	resp, err := callAPI(testServer.URL, &user, "", startTime, endTime, 0, 100, "", 123)
	if err == nil || resp != nil {
		t.Fail()
	}
}

func TestCallAPIWithInactiveSession(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: false}
	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	call("http://localhost:8080/v1", &APIConf{API: "/pmdata"}, &user, "", startTime, endTime, 123)
	if !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestStoreLastReceivedDataTime(t *testing.T) {
	resp := new(GetAPIResponse)
	err := json.NewDecoder(bytes.NewReader([]byte(response))).Decode(resp)

	user := &User{Email: "testuser@nokia.com", ResponseDest: "./tmp", isSessionAlive: true}
	err = storeLastReceivedDataTime(user, "/fm", resp.Data, fmResponseType, "ACTIVE", "", 123)
	if err != nil {
		t.Fail()
	}
	//Reading LastReceivedFile value from file
	fileName := "fm" + "_" + "ACTIVE" + "_" + user.Email
	data, err := ioutil.ReadFile(fileName)
	defer os.Remove(fileName)
	if err != nil && len(data) == 0 {
		t.Fail()
	}
}

func TestStoreLastReceivedDataTimeWithoutData(t *testing.T) {
	responseDir := "./tmp"
	if _, err := os.Stat(responseDir); os.IsNotExist(err) {
		os.MkdirAll(responseDir, os.ModePerm)
	}
	defer os.RemoveAll(responseDir)
	user := &User{Email: "testuser@nokia.com", ResponseDest: "./tmp", isSessionAlive: true}
	err := storeLastReceivedDataTime(user, "", "", fmResponseType, "ACTIVE", "test_nhg", 123)
	t.Log(err)
	if err == nil || !strings.Contains(err.Error(), "Unable to write last received data time, error: unexpected end of JSON input") {
		t.Fail()
	}
}

func TestStoreLastReceivedFileTimeWithWrongDirectory(t *testing.T) {
	user := &User{Email: "testuser@nokia.com", ResponseDest: "./tmp", isSessionAlive: true}
	err := storeLastReceivedDataTime(user, "", "", fmResponseType, "ACTIVE", "test_nhg", 123)
	t.Log(err)
	if err == nil {
		t.Fail()
	}
}

func TestCreateResponseDirectory(t *testing.T) {
	respDir := "./tmp"
	CreateResponseDirectory(respDir, "http://localhost:8080/pmdata")
	defer os.RemoveAll(respDir)
	if _, err := os.Stat(respDir + "/pmdata"); os.IsNotExist(err) {
		t.Fail()
	}
}

func createTmpFile(dir string, prefix string, content []byte) string {
	tmpfile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		log.Error(err)
		return ""
	}

	if _, err = tmpfile.Write(content); err != nil {
		log.Error(err)
		return ""
	}
	if err = tmpfile.Close(); err != nil {
		log.Error(err)
		return ""
	}
	return tmpfile.Name()
}

func TestStartDataCollectionWithInvalidURL(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	oldCurrentTime := currentTime
	defer func() { currentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	currentTime = myCurrentTime
	Conf.BaseURL = "http://localhost:8080"
	Conf.Users = []*User{
		{
			Email:          "testuser@nokia.com",
			Password:       "MTIzNA==",
			isSessionAlive: true,
			ResponseDest:   "./tmp",
			sessionToken: &sessionToken{
				accessToken:  "",
				refreshToken: "",
				expiryTime:   currentTime(),
			},
		},
	}
	Conf.APIs = []*APIConf{{API: "/pmdata", Interval: 15}, {API: "/fmdata", Interval: 15, Type: "HISTORY"}}
	Conf.Limit = 10
	CreateHTTPClient("", true)

	StartDataCollection()
	time.Sleep(2 * time.Millisecond)
	if !strings.Contains(buf.String(), "Triggered http://localhost:8080/fmdata") || !strings.Contains(buf.String(), "Triggered http://localhost:8080/pmdata") {
		t.Fail()
	}
	if strings.Contains(buf.String(), "Writting response") {
		t.Fail()
	}
}

func TestStartDataCollection(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, response)
	}))
	defer testServer.Close()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	oldCurrentTime := currentTime
	defer func() { currentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	currentTime = myCurrentTime
	Conf.BaseURL = testServer.URL
	Conf.Users = []*User{
		{
			Email:          "testuser@nokia.com",
			Password:       "MTIzNA==",
			ResponseDest:   "./tmp",
			isSessionAlive: true,
			sessionToken: &sessionToken{
				accessToken:  "",
				refreshToken: "",
				expiryTime:   currentTime(),
			},
		},
	}
	Conf.APIs = []*APIConf{{API: "/pmdata", Interval: 15}, {API: "/fmdata", Interval: 15, Type: "HISTORY"}}
	Conf.Limit = 10
	CreateHTTPClient("", true)

	StartDataCollection()
	time.Sleep(2 * time.Millisecond)
	if !strings.Contains(buf.String(), "Triggered "+testServer.URL) {
		t.Fail()
	}
	if !strings.Contains(buf.String(), "Writing response") {
		t.Fail()
	}
}

func TestAPICallWithPagination(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		index, _ := strconv.Atoi(query.Get(indexQueryParam))
		w.Header().Set("Content-Type", "application/json")
		if index == 0 {
			fmt.Fprintln(w, response1)
		} else {
			fmt.Fprintln(w, response2)
		}
	}))
	defer testServer.Close()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	CreateHTTPClient("", true)

	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 1000)
	call(testServer.URL, &APIConf{API: "/fmdata", Interval: 15}, &user, "", startTime, endTime, 1000)
	if !strings.Contains(buf.String(), "\"Received response details\" nhg_id= received_no_of_records=4 tid=1000 total_no_of_records=4") {
		t.Error(buf.String())
	}
}

func TestRetryAPICall(t *testing.T) {
	count := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		index, _ := strconv.Atoi(query.Get(indexQueryParam))
		w.Header().Set("Content-Type", "application/json")
		if index == 0 {
			fmt.Fprintln(w, response1)
		} else if count == 2 {
			fmt.Fprintln(w, response2)
		} else {
			count++
			fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Token sent is empty. Invalid Token"
				}
			}
		}`)
		}
	}))
	defer testServer.Close()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	CreateHTTPClient("", true)

	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 123)
	call(testServer.URL, &APIConf{API: "/fmdata", Interval: 15}, &user, "", startTime, endTime, 1000)
	if !strings.Contains(buf.String(), "\"Received response details\" nhg_id= received_no_of_records=4 tid=1000 total_no_of_records=4") {
		t.Error(buf.String())
	}
}

func TestRetryAPICallFromStarting(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		index, _ := strconv.Atoi(query.Get(indexQueryParam))
		w.Header().Set("Content-Type", "application/json")
		if index == 0 {
			fmt.Fprintln(w, response1)
		} else {
			fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Token sent is empty. Invalid Token"
				}
			}
		}`)
		}
	}))
	defer testServer.Close()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	CreateHTTPClient("", true)

	startTime, endTime := getTimeInterval(&user, &APIConf{API: "/fmdata", Interval: 15}, "", 1000)
	call(testServer.URL, &APIConf{API: "/fmdata", Interval: 15}, &user, "", startTime, endTime, 1000)
	if !strings.Contains(buf.String(), "API call failed, data will be skipped for the duration") {
		t.Error(buf.String())
	}
}

func TestGetNhgDetails(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listNhgResp)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	getNhgDetails(testServer.URL, &APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234)
	if user.nhgIDs[0] != "test_nhg_2" {
		t.Fail()
	}
}

func TestGetNhgDetailsWithInactiveSession(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: false}
	getNhgDetails("http://localhost:8080/v1", &APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234)
	if !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestGetNhgDetailsForInvalidCase(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Token sent is empty. Invalid Token"
				}
			}
		}`)
	}))
	defer testServer.Close()

	CreateHTTPClient("", true)
	getNhgDetails(testServer.URL, &APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234)
	if len(user.nhgIDs) != 0 {
		t.Fail()
	}
}

func TestGetNhgDetailsForInvalidURL(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}

	CreateHTTPClient("", true)
	getNhgDetails(":", &APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234)
	if !strings.Contains(buf.String(), "missing protocol scheme") {
		t.Fail()
	}
}

func TestGetNhgDetailsWithInvalidResponse(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA==", isSessionAlive: true}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   currentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	CreateHTTPClient("", true)
	getNhgDetails(testServer.URL, &APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234)
	if !strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
}
