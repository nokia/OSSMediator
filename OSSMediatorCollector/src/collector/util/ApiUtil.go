/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	//LastReceivedDataTime keeps the timestamp of last file received from the GET API call
	//It will be sent as query parameter to the API to send response after the LastReceivedDataTime timestamp.
	lastReceivedDataTime = "./LastReceivedDataTime"

	//Timeout duration for HTTP calls
	timeout = 62 * time.Second

	//Maximum no. of retry attempt for API call
	maxAttempts = 3

	//Time interval at which the 1st API call should start
	interval = 15

	//Query Params for GET APIS
	startTimeQueryParam = "start_timestamp"
	endTimeQueryParam   = "end_timestamp"
	limitQueryParam     = "limit"
	indexQueryParam     = "index"
	alarmTypeQueryParam = "alarm_type"
	nhgIDQueryParam     = "nhg_id"

	//Headers
	authorizationHeader = "Authorization"

	//Success status code from response
	successStatusCode = "SUCCESS"

	//Backoff duration for retrying refresh token.
	initialBackoff = 5 * time.Second
	maxBackoff     = 120 * time.Second
	multiplier     = 2

	//field name to extract data from PM response file
	sourceField      = "_source"
	eventTimeFieldPM = "timestamp"
	eventTimeFieldFM = "EventTime"

	//event time format
	eventTimeFormatBaiCell   = "2006-01-02T15:04:05Z07:00"
	eventTimeFormatNokiaCell = "2006-01-02T15:04:05Z07:00:00"

	//File extension for writing response
	fileExtension = ".json"

	//response type
	fmResponseType = "fmdata"
	pmResponseType = "pmdata"
)

var (
	//Conf keeps the config from json and console
	Conf Config

	//HTTP client for all API calls
	client *http.Client

	//used for logging
	txnID uint64 = 1000
)

//GetAPIResponse keeps track of response received from PM/FM API
type GetAPIResponse struct {
	Type            string `json:"type"`
	TotalNumRecords int    `json:"total_num_records"`
	NumOfRecords    int    `json:"num_of_records"`
	NextRecord      int    `json:"next_record"`
	Data            string `json:"data"`
	Status          Status `json:"status"` // Status of the response
}

//NhgDetailsResponse struct for listNhgData API
type NhgDetailsResponse struct {
	Status     Status      `json:"status"` // Status of the response
	NhgDetails []NhgDetail `json:"nhg_info"`
}

//NhgDetail stores nhg details
type NhgDetail struct {
	NhgID    string `json:"nhg_id"`
	NwStatus string `json:"nhg_config_status"`
}

//Status keeps track of status from response
type Status struct {
	StatusCode        string            `json:"status_code"`
	StatusDescription StatusDescription `json:"status_description"`
}

//StatusDescription keeps track of status description from response
type StatusDescription struct {
	DescriptionCode string `json:"description_code"`
	Description     string `json:"description"`
}

//sessionToken struct tracks the access_token, refresh_token and expiry_time of the token
//As the session token will be shared by multiple APIs.
type sessionToken struct {
	access       sync.Mutex
	accessToken  string
	refreshToken string
	expiryTime   time.Time
}

type fn func(string, *APIConf, *User, uint64)

var currentTime = time.Now

//StartDataCollection starts the tickers for PM/FM APIs.
func StartDataCollection() {
	currentTime := currentTime()
	diff := currentTime.Minute() - (currentTime.Minute() / interval * interval) - Conf.Delay
	begTime := currentTime.Add(time.Duration(-1*diff) * time.Minute)
	if currentTime.After(begTime) {
		begTime = begTime.Add(time.Duration(interval) * time.Minute)
		for _, user := range Conf.Users {
			if Conf.ListNhGAPI != nil {
				getNhgDetails(Conf.BaseURL, Conf.ListNhGAPI, user, atomic.AddUint64(&txnID, 1))
			}
			for _, api := range Conf.APIs {
				go fetchMetricsData(Conf.BaseURL, api, user, atomic.AddUint64(&txnID, 1))
			}
		}
	}

	timer := time.NewTimer(time.Until(begTime))
	<-timer.C
	//For each APIs creates ticker to trigger the API periodically at specified interval.
	for _, user := range Conf.Users {
		if Conf.ListNhGAPI != nil {
			getNhgDetails(Conf.BaseURL, Conf.ListNhGAPI, user, atomic.AddUint64(&txnID, 1))
			ticker := time.NewTicker(time.Duration(Conf.ListNhGAPI.Interval) * time.Minute)
			go trigger(ticker, Conf.BaseURL, Conf.ListNhGAPI, user, getNhgDetails)
		}
		for _, api := range Conf.APIs {
			go fetchMetricsData(Conf.BaseURL, api, user, atomic.AddUint64(&txnID, 1))
			ticker := time.NewTicker(time.Duration(api.Interval) * time.Minute)
			go trigger(ticker, Conf.BaseURL, api, user, fetchMetricsData)
		}
	}
}

func fetchMetricsData(baseURL string, api *APIConf, user *User, txnID uint64) {
	if len(user.nhgIDs) == 0 {
		startTime, endTime := getTimeInterval(user, api, "", txnID)
		call(baseURL, api, user, "", startTime, endTime, txnID)
	} else {
		for _, nhgID := range user.nhgIDs {
			startTime, endTime := getTimeInterval(user, api, nhgID, txnID)
			call(baseURL, api, user, nhgID, startTime, endTime, txnID)
		}
	}
}

//calls the API and handles pagination requests
func call(baseURL string, api *APIConf, user *User, nhgID string, startTime string, endTime string, txnID uint64) {
	apiURL := baseURL + api.API
	if !user.isSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api_url": apiURL, "nhg_id": nhgID}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, currentTime())
		return
	}

	log.WithFields(log.Fields{"tid": txnID, "nhg_id": nhgID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, currentTime())
	response, err := callAPI(apiURL, user, nhgID, startTime, endTime, 0, Conf.Limit, api.Type, txnID)
	if err != nil {
		if !strings.Contains(err.Error(), "no records found") {
			response, err = retryAPICall(apiURL, user, nhgID, startTime, endTime, 0, Conf.Limit, api.Type, txnID)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "nhg_id": nhgID, "api_url": apiURL, "start_time": startTime, "end_time": endTime}).Infof("API call failed, data will be skipped...")
				return
			}
		} else {
			return
		}
	}
	receivedNoOfRecords := response.NumOfRecords
	totalNoOfRecords := response.TotalNumRecords
	nextRecord := response.NextRecord
	retryCompleteCallFlag := false

	//handling pagination
	for err == nil && nextRecord > 0 {
		response, err = callAPI(apiURL, user, nhgID, startTime, endTime, nextRecord, Conf.Limit, api.Type, txnID)
		if err != nil {
			response, err = retryAPICall(apiURL, user, nhgID, startTime, endTime, nextRecord, Conf.Limit, api.Type, txnID)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "nhg_id": nhgID, "api_url": apiURL, "start_time": startTime, "end_time": endTime}).Infof("API call failed, will be retried from starting...")
				retryCompleteCallFlag = true
			} else {
				receivedNoOfRecords += response.NumOfRecords
				nextRecord = response.NextRecord
			}
		} else {
			receivedNoOfRecords += response.NumOfRecords
			nextRecord = response.NextRecord
		}
	}

	log.WithFields(log.Fields{"tid": txnID, "nhg_id": nhgID, "total_no_of_records": totalNoOfRecords, "received_no_of_records": receivedNoOfRecords}).Infof("Received response details")
	if totalNoOfRecords > receivedNoOfRecords {
		log.WithFields(log.Fields{"tid": txnID, "nhg_id": nhgID, "total_no_of_records": totalNoOfRecords, "received_no_of_records": receivedNoOfRecords}).Warnf("Received no. of records less than total no. of records, API call will be retried from starting...")
		retryCompleteCallFlag = true
	}

	//retrying API call from beginning
	var retryResponse *GetAPIResponse
	if retryCompleteCallFlag {
		retryResponse, err = callAPI(apiURL, user, nhgID, startTime, endTime, 0, Conf.Limit, api.Type, txnID)
		if err != nil {
			log.WithFields(log.Fields{"txn_id": txnID, "api_url": apiURL, "start_time": startTime, "end_time": endTime, "index": nextRecord, "nhg_id": nhgID}).Warnf("API call failed, data will be skipped for the duration")
		} else {
			nextRecord = retryResponse.NextRecord
			for nextRecord > 0 {
				retryResponse, err = callAPI(apiURL, user, nhgID, startTime, endTime, nextRecord, Conf.Limit, api.Type, txnID)
				if err != nil {
					log.WithFields(log.Fields{"txn_id": txnID, "api_url": apiURL, "start_time": startTime, "end_time": endTime, "index": nextRecord, "nhg_id": nhgID}).Warnf("API call failed, data will be skipped for the duration")
					break
				}
				nextRecord = retryResponse.NextRecord
			}
		}
	}
}

//retry failed API call 3 times
func retryAPICall(apiURL string, user *User, nhgID string, startTime string, endTime string, indx int, limit int, apiType string, txnID uint64) (*GetAPIResponse, error) {
	var err error
	var response *GetAPIResponse
	for i := 0; i < maxAttempts; i++ {
		log.WithFields(log.Fields{"txn_id": txnID, "nhg_id": nhgID, "api_url": apiURL, "start_time": startTime, "end_time": endTime, "limit": limit, "index": indx}).Info("retrying api call")
		response, err = callAPI(apiURL, user, nhgID, startTime, endTime, indx, limit, apiType, txnID)
		if err == nil {
			return response, nil
		}
	}
	return nil, err
}

//Trigger the method periodically at specified interval.
func trigger(ticker *time.Ticker, baseURL string, api *APIConf, user *User, method fn) {
	for {
		<-ticker.C
		method(baseURL, api, user, atomic.AddUint64(&txnID, 1))
	}
}

//CreateHTTPClient creates HTTP client for all the GET/POST API calls, if certFile is empty and skipTLS is false TLS authentication will be done using root certificates.
//certFile keeps the server certificate file path
//skipTLS if true all API calls will skip TLS auth.
func CreateHTTPClient(certFile string, skipTLS bool) {
	if skipTLS {
		//skipping certificates
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr, Timeout: timeout}
		log.Debugf("Skipping TLS authentication")
	} else if certFile == "" {
		client = &http.Client{Timeout: timeout}
		log.Debugf("TLS authentication using root certificates")
	} else {
		//Load CA cert
		caCert, err := ioutil.ReadFile(certFile)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Error while reading server certificate file")
			client = &http.Client{Timeout: timeout}
			log.Debugf("TLS authentication using root certificates")
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		//Setup HTTPS client
		tlsConfig := &tls.Config{
			RootCAs: caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		tr := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: tr, Timeout: timeout}
		log.Debugf("Using CA certificate %s", certFile)
	}
}

//CallAPI calls the API, adds authorization, query params and returns response.
//If successful it returns response as array of byte, if there is any error it return nil.
func callAPI(apiURL string, user *User, nhgID string, startTime string, endTime string, indx int, limit int, apiType string, txnID uint64) (*GetAPIResponse, error) {
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return nil, err
	}
	//Lock the SessionToken object before calling API
	user.sessionToken.access.Lock()
	defer user.sessionToken.access.Unlock()
	request.Header.Set(authorizationHeader, user.sessionToken.accessToken)

	//requesting compressed response
	request.Header.Add("Accept-Encoding", "gzip")

	//Adding query params
	query := request.URL.Query()
	query.Add(startTimeQueryParam, startTime)
	query.Add(endTimeQueryParam, endTime)
	query.Add(limitQueryParam, strconv.Itoa(limit))
	query.Add(indexQueryParam, strconv.Itoa(indx))
	if apiType != "" {
		query.Add(alarmTypeQueryParam, apiType)
	}
	if nhgID != "" {
		query.Add(nhgIDQueryParam, nhgID)
	}
	request.URL.RawQuery = query.Encode()
	log.WithFields(log.Fields{"tid": txnID}).Info(startTimeQueryParam, ": ", query[startTimeQueryParam])
	log.WithFields(log.Fields{"tid": txnID}).Info(endTimeQueryParam, ": ", query[endTimeQueryParam])
	log.WithFields(log.Fields{"tid": txnID}).Info("URL:", request.URL)

	response, err := doRequest(request)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": nhgID, "api_url": apiURL, "start_time": startTime, "end_time": endTime}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return nil, err
	}

	//Map the received response to getAPIResponse struct
	resp := new(GetAPIResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": nhgID, "api_url": apiURL, "start_time": startTime, "end_time": endTime}).Error("Unable to decode response")
		return nil, err
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": nhgID, "api_url": apiURL, "start_time": startTime, "end_time": endTime}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return nil, err
	}
	log.WithFields(log.Fields{"tid": txnID, "nhg_id": nhgID, "api_url": apiURL, "start_time": startTime, "end_time": endTime, "total_no_of_records": resp.TotalNumRecords, "no_of_records_received": resp.NumOfRecords, "next_record_index": resp.NextRecord}).Infof("%s called successfully for %s.", apiURL, user.Email)
	log.WithFields(log.Fields{"tid": txnID}).Debugf("response: %v", resp)

	//storing LastReceivedDataTime timestamp value to file
	err = storeLastReceivedDataTime(user, apiURL, resp.Data, resp.Type, apiType, nhgID, txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID}).Error(err)
	}

	//write response
	writeResponse(resp, user, apiURL, txnID)
	return resp, nil
}

//Executes the request.
//If successful returns response and nil, if there is any error it return error.
func doRequest(request *http.Request) ([]byte, error) {
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if 200 != response.StatusCode {
		return nil, fmt.Errorf("%d: %s", response.StatusCode, response.Status)
	}

	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	default:
		reader = response.Body
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, nil
}

//Validates the response's status code.
//If status code = SUCCESS it returns nil, else it returns invalid status code error.
func checkStatusCode(status Status) error {
	if status.StatusCode != successStatusCode {
		return fmt.Errorf("Error while validating response status: Status Code: %s, Status Message: %s", status.StatusCode, status.StatusDescription.Description)
	}
	return nil
}

//writeResponse reads the response and writes the data in json format to responseDest directory.
//Return index of next record to be fetched.
func writeResponse(response *GetAPIResponse, user *User, apiURL string, txnID uint64) {
	log.WithFields(log.Fields{"tid": txnID}).Infof("Writing response for %s to %s", user.Email, user.ResponseDest+"/"+path.Base(apiURL))
	fileName := response.Type + "_response_" + strconv.Itoa(int(currentTime().Unix()))
	responseDest := user.ResponseDest + "/" + path.Base(apiURL)
	fileName = responseDest + "/" + fileName
	counter := 1
	name := fileName
	for fileExists(name + fileExtension) {
		name = fileName + "_" + strconv.Itoa(counter)
		counter++
	}
	fileName = name + fileExtension
	var out bytes.Buffer
	err := json.Indent(&out, []byte(response.Data), "", "  ")
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to indent received data, error: %v", err)
	}

	log.WithFields(log.Fields{"tid": txnID}).Infof("Writing response to file %s for %s", fileName, user.Email)
	err = writeFile(fileName, out.Bytes())
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Writing response to file %s for %s failed", fileName, user.Email)
	}
}

//checks whether file exists, return false if file ids not present.
func fileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

//writes data to file, returns error if file creation or data writing failed.
func writeFile(fileName string, data []byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("File creation failed: %v", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Error while writting response to file: %v", err)
	}
	return nil
}

//Writes the last received metric's event time to a file so that next time that time stamp will be used as start_time for api calls.
//Creates file for each user and each APIs.
//returns error if writing to or reading from the response directory fails.
func storeLastReceivedDataTime(user *User, apiURL string, data string, respType string, apiType string, nhgID string, txnID uint64) error {
	var metricData interface{}
	err := json.Unmarshal([]byte(data), &metricData)
	if err != nil {
		return fmt.Errorf("Unable to write last received data time, error: %v", err)
	}

	var fieldName string
	if respType == fmResponseType {
		fieldName = eventTimeFieldFM
	} else if respType == pmResponseType {
		fieldName = eventTimeFieldPM
	}

	receivedData := make(map[string]struct{})
	for _, value := range metricData.([]interface{}) {
		metricTime := value.(map[string]interface{})[sourceField].(map[string]interface{})[fieldName].(string)
		receivedData[metricTime] = struct{}{}
	}
	eventTimes := []time.Time{}
	for key := range receivedData {
		eventTime, err := time.Parse(eventTimeFormatBaiCell, key)
		if err != nil {
			eventTime, err = time.Parse(eventTimeFormatNokiaCell, key)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "event_time": eventTime}).Errorf("Unable to parse event time using Nokia cell or BaiCell time format")
				continue
			}
		}
		eventTimes = append(eventTimes, eventTime)
	}
	lastReceivedTime := getLastReceivedDataTime(user, apiURL, apiType, nhgID, txnID)
	if lastReceivedTime != "" {
		t, err := time.Parse(time.RFC3339, lastReceivedTime)
		if err == nil {
			eventTimes = append(eventTimes, t)
		}
	}
	sort.Slice(eventTimes, func(i, j int) bool { return eventTimes[i].Before(eventTimes[j]) })
	latestTime := truncateSeconds(eventTimes[len(eventTimes)-1]).Format(time.RFC3339)

	fileName := path.Base(apiURL)
	if apiType != "" {
		fileName = fileName + "_" + apiType
	}
	fileName = fileName + "_" + user.Email
	if nhgID != "" {
		fileName = fileName + "_" + nhgID
	}
	log.WithFields(log.Fields{"tid": txnID}).Debug("Writing to ", fileName)
	err = writeFile(fileName, []byte(latestTime))
	if err != nil {
		return fmt.Errorf("Unable to write last received data time, error: %v", err)
	}
	return nil
}

//CreateResponseDirectory creates directory named path, along with any necessary parents.
// If the directory creation fails it will terminate the program.
func CreateResponseDirectory(basePath string, api string) {
	path := basePath + "/" + path.Base(api)
	log.Infof("Creating %s directory", path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalf("Error while creating %s", path)
	}
}

//truncates seconds from time
func truncateSeconds(t time.Time) time.Time {
	return t.Truncate(60 * time.Second)
}

//getTimeInterval: returns the start_time and end_time for API calls.
//It reads start_time from file where last received data's event time is stored, if file is not present or
func getTimeInterval(user *User, api *APIConf, nhgID string, txnID uint64) (string, string) {
	currentTime := currentTime().Add(time.Duration(-10) * time.Second)
	//calculating 15 minutes time frame
	diff := currentTime.Minute() - (currentTime.Minute() / api.Interval * api.Interval) + api.Interval
	begTime := currentTime.Add(time.Duration(-1*diff) * time.Minute)
	endTime := begTime.Add(time.Duration(api.Interval) * time.Minute)
	startTime := getLastReceivedDataTime(user, api.API, api.Type, nhgID, txnID)
	if startTime == "" {
		startTime = truncateSeconds(begTime).Format(time.RFC3339)
	}
	if api.SyncDuration > 0 {
		stTime, _ := time.Parse(time.RFC3339, startTime)
		gap := endTime.Sub(stTime)
		if gap < (time.Duration(api.SyncDuration) * time.Minute) {
			stTime = endTime.Add(time.Duration(-1*api.SyncDuration) * time.Minute)
			startTime = truncateSeconds(stTime).Format(time.RFC3339)
		}
	}

	return startTime, endTime.Format(time.RFC3339)

}

//retrieves the last received metric time from file for per API
func getLastReceivedDataTime(user *User, apiURL string, apiType string, nhgID string, txnID uint64) string {
	//Reading start time value from file
	fileName := path.Base(apiURL)
	if apiType != "" {
		fileName = fileName + "_" + apiType
	}
	fileName = fileName + "_" + user.Email
	if nhgID != "" {
		fileName = fileName + "_" + nhgID
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil || len(data) == 0 {
		return ""
	}
	return strings.TrimSpace(string(data))
}

//get nhg details for the customer
func getNhgDetails(baseURL string, api *APIConf, user *User, txnID uint64) {
	apiURL := baseURL + api.API
	if !user.isSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api_url": apiURL}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, currentTime())
		return
	}
	log.WithFields(log.Fields{"tid": txnID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, currentTime())
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}
	//Lock the SessionToken object before calling API
	user.sessionToken.access.Lock()
	request.Header.Set(authorizationHeader, user.sessionToken.accessToken)
	user.sessionToken.access.Unlock()

	response, err := doRequest(request)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	//Map the received response to NhgDetailsResponse struct
	resp := new(NhgDetailsResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return
	}
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "nhg_ids": resp.NhgDetails}).Info("Response")

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return
	}

	user.nhgIDs = []string{}
	for _, nhgInfo := range resp.NhgDetails {
		if nhgInfo.NwStatus == "ACTIVE" {
			user.nhgIDs = append(user.nhgIDs, nhgInfo.NhgID)
		}
	}
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "nhg_ids": user.nhgIDs}).Info("user nhgs")
}
