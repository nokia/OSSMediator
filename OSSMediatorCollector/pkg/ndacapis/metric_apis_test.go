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
)

var (
	fmResponse = `{
		"data": [{"fm_data":{"alarm_identifier":"2","severity":"major","specific_problem":"111","alarm_text":"Synchronizationlost","additional_text":"Synchronizationlost","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}},{"fm_data":{"alarm_identifier":"3","severity":"minor","specific_problem":"2222","alarm_text":"SSHenabled","additional_text":"SSHenabled","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}}],
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

	paginationResponse1 = `{
		"data": [{"fm_data":{"alarm_identifier":"2","severity":"major","specific_problem":"111","alarm_text":"Synchronizationlost","additional_text":"Synchronizationlost","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}},{"fm_data":{"alarm_identifier":"3","severity":"minor","specific_problem":"2222","alarm_text":"SSHenabled","additional_text":"SSHenabled","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}}],
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

	paginationResponse2 = `{
		"data": [{"fm_data":{"alarm_identifier":"2","severity":"major","specific_problem":"111","alarm_text":"Synchronizationlost","additional_text":"Synchronizationlost","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}},{"fm_data":{"alarm_identifier":"3","severity":"minor","specific_problem":"2222","alarm_text":"SSHenabled","additional_text":"SSHenabled","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}}],
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

	largeResponse = `{
		"data": [{"fm_data":{"alarm_identifier":"2","severity":"major","specific_problem":"111","alarm_text":"Synchronizationlost","additional_text":"Synchronizationlost","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}},{"fm_data":{"alarm_identifier":"3","severity":"minor","specific_problem":"2222","alarm_text":"SSHenabled","additional_text":"SSHenabled","alarm_state":"ACTIVE","event_type":"communications","event_time":"2020-10-30T13:29:27Z","last_updated_time":"2020-11-05T08:10:36Z","notification_type":"alarmNew"},"fm_data_source":{"hw_alias":"testhw","serial_no":"12345","nhg_id":"test_nhg_1","nhg_alias":"testnhg"}}],
		"next_record": 10000,
		"num_of_records": 10000,
		"status": {
		  "status_code": "SUCCESS",
		  "status_description": {
			"description": "Records transferred successfully",
			"description_code": "NOT_SPECIFIED"
		  }
		},
		"total_num_records": 20000,
		"type": "fmdata"
	  }`
)

func TestCallAPIForInvalidCase(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
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
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}

	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		url:       testServer.URL + apiConf.API,
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}
	response, err := callAPI(apiReq, 123)
	if err == nil || response != nil || !strings.Contains(err.Error(), "error while validating response status") {
		t.Fail()
	}
}

func TestCallAPIForInvalidURL(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}

	CreateHTTPClient("", true)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		url:       ":" + apiConf.API,
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}
	response, err := callAPI(apiReq, 123)
	if err == nil || response != nil || !strings.Contains(err.Error(), "missing protocol scheme") {
		t.Fail()
	}
}

func TestCallAPI(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fmResponse)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	defer os.RemoveAll(user.ResponseDest)

	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		url:       testServer.URL + apiConf.API,
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}
	resp, err := callAPI(apiReq, 123)
	if err != nil || resp.Status.StatusCode != "SUCCESS" || resp.Type != "fmdata" || resp.TotalNumRecords != 2 || resp.NumOfRecords != 2 || resp.NextRecord != 0 {
		t.Fail()
	}
}

func TestCallAPIWithInvalidResponse(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}

	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		url:       testServer.URL + apiConf.API,
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}

	resp, err := callAPI(apiReq, 123)
	if resp != nil && strings.Contains(err.Error(), "Unable to decode response") {
		t.Fail()
	}
}

func TestCallAPIWithSkipCert(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fmResponse)
	}))
	defer testServer.Close()
	CreateHTTPClient("", true)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	defer os.RemoveAll(user.ResponseDest)

	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		url:       testServer.URL + apiConf.API,
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}
	resp, err := callAPI(apiReq, 123)
	if err != nil || resp.Status.StatusCode != "SUCCESS" || resp.Type != "fmdata" || resp.TotalNumRecords != 2 || resp.NumOfRecords != 2 || resp.NextRecord != 0 {
		t.Fail()
	}
}

func TestCallAPIWithCert(t *testing.T) {
	tmpfile := createTmpFile(".", "crt", []byte(``))
	CreateHTTPClient(tmpfile, false)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fmResponse)
	}))
	defer testServer.Close()
	defer os.Remove(tmpfile)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	defer os.RemoveAll(user.ResponseDest)

	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		url:       testServer.URL + apiConf.API,
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}

	resp, err := callAPI(apiReq, 123)
	if err != nil || resp.Status.StatusCode != "SUCCESS" || resp.Type != "fmdata" || resp.TotalNumRecords != 2 || resp.NumOfRecords != 2 || resp.NextRecord != 0 {
		t.Fail()
	}
}

func TestCallAPIWithErrorStatusCode(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		url:       testServer.URL + apiConf.API,
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}
	resp, err := callAPI(apiReq, 123)
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
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: false}
	api := &config.APIConf{API: "/fmdata", Interval: 15}
	fetchMetricsData(api, &user, 123)

	if !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestAPICallWithPagination(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		index, _ := strconv.Atoi(query.Get(indexQueryParam))
		w.Header().Set("Content-Type", "application/json")
		if index == 0 {
			fmt.Fprintln(w, paginationResponse1)
		} else {
			fmt.Fprintln(w, paginationResponse2)
		}
	}))
	defer testServer.Close()

	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	m := map[string]config.OrgAccDetails{}
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails.OrgUUID = "org_uuid_1"
	orgAcc.OrgDetails.OrgAlias = "org_alias_1"
	orgAcc.AccDetails.AccUUID = "acc_uuid_1"
	orgAcc.AccDetails.AccAlias = "acc_alias_1"

	m["test_nhg_1"] = orgAcc
	user.NhgIDs = m
	CreateHTTPClient("", true)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	config.Conf.BaseURL = testServer.URL
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)

	fetchMetricsData(apiConf, &user, 123)
	files, err := ioutil.ReadDir(user.ResponseDest + apiConf.API)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Fail()
	}
	os.RemoveAll(user.ResponseDest)
}

func TestRetryAPICall(t *testing.T) {
	count := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		index, _ := strconv.Atoi(query.Get(indexQueryParam))
		w.Header().Set("Content-Type", "application/json")
		if index == 0 {
			fmt.Fprintln(w, paginationResponse1)
		} else if count == 2 {
			fmt.Fprintln(w, paginationResponse2)
		} else {
			count++
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer testServer.Close()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}

	m := map[string]config.OrgAccDetails{}
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails.OrgUUID = "org_uuid_1"
	orgAcc.OrgDetails.OrgAlias = "org_alias_1"
	orgAcc.AccDetails.AccUUID = "acc_uuid_1"
	orgAcc.AccDetails.AccAlias = "acc_alias_1"

	m["test_nhg_1"] = orgAcc
	user.NhgIDs = m
	CreateHTTPClient("", true)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	config.Conf.BaseURL = testServer.URL
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)

	fetchMetricsData(apiConf, &user, 123)
	files, err := ioutil.ReadDir(user.ResponseDest + apiConf.API)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Fail()
	}
	os.RemoveAll(user.ResponseDest)
}

func TestCallMetricAPIForLargeResponse(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	count := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if count == 0 {
			fmt.Fprintln(w, largeResponse)
		} else if count == 1 {
			fmt.Fprintln(w, fmResponse)
		}
		count++
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)
	defer os.RemoveAll(user.ResponseDest)

	startTime, endTime := utils.GetTimeInterval(&user, apiConf, "")
	if startTime == "" || endTime == "" {
		t.Fail()
	}
	apiReq := apiCallRequest{
		api:       apiConf,
		user:      &user,
		nhgID:     "",
		startTime: startTime,
		endTime:   endTime,
		index:     0,
		limit:     100,
	}
	config.Conf.BaseURL = testServer.URL

	status := callMetricAPI(apiReq, 1, 123)
	if status != "" {
		t.Fail()
	}
}

func TestFetchMetricsDataWithRetryNext(t *testing.T) {
	count := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if count == 0 {
			fmt.Fprintln(w, largeResponse)
		} else if count == 1 {
			fmt.Fprintln(w, fmResponse)
		}
		count++
	}))
	defer testServer.Close()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	m := map[string]config.OrgAccDetails{}
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails.OrgUUID = "org_uuid_1"
	orgAcc.OrgDetails.OrgAlias = "org_alias_1"
	orgAcc.AccDetails.AccUUID = "acc_uuid_1"
	orgAcc.AccDetails.AccAlias = "acc_alias_1"

	m["test_nhg_1"] = orgAcc
	user.NhgIDs = m

	CreateHTTPClient("", true)
	apiConf := &config.APIConf{API: "/fmdata", Interval: 15}
	config.Conf.BaseURL = testServer.URL
	utils.CreateResponseDirectory(user.ResponseDest, apiConf.API)

	fetchMetricsData(apiConf, &user, 123)
	files, err := ioutil.ReadDir(user.ResponseDest + apiConf.API)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Fail()
	}
	os.RemoveAll(user.ResponseDest)
}
