/*
* Copyright 2018 Nokia 
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
*/

package util

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	//LastReceivedDataTime keeps the timestamp of last file received from the GET API call
	//It will be sent as query parameter to the API to send response after the LastReceivedDataTime timestamp.
	lastReceivedDataTime = "./LastReceivedDataTime"

	//Timeout duration for HTTP calls
	timeout = time.Duration(62 * time.Second)

	//Query Params for GET APIS
	startTimeQueryParam = "start_timestamp"
	endTimeQueryParam   = "end_timestamp"
	limitQueryParam     = "limit"
	indexQueryParam     = "index"
	alarmTypeQueryParam = "alarm_type"

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

var currentTime = time.Now

func StartDataCollection() {
	interval := 15
	currentTime := currentTime()
	diff := currentTime.Minute() - (currentTime.Minute() / interval * interval) - 7
	begTime := currentTime.Add(time.Duration(-1*diff) * time.Minute)
	if currentTime.After(begTime) {
		begTime = begTime.Add(time.Duration(15)*time.Minute)
		for _, user := range Conf.Users {
			for _, api := range Conf.APIs {
				call(Conf.BaseURL, api, user)
			}
		}
	}

	timer := time.NewTimer(time.Until(begTime))
	<-timer.C
	//For each APIs creates ticker to trigger the API periodically at specified interval.
	for _, user := range Conf.Users {
		for _, api := range Conf.APIs {
			call(Conf.BaseURL, api, user)
			ticker := time.NewTicker(time.Duration(api.Interval) * time.Minute)
			go trigger(ticker, Conf.BaseURL, api, user)
		}
	}
}

func call(baseURL string, api APIConf, user *User) {
	apiURL := baseURL + api.API
	log.Infof("Triggered %s for %s at %v", apiURL, user.Email, currentTime())
	startTime, endTime := getTimeInterval(user, apiURL, api.Interval)
	response := callAPI(apiURL, user, startTime, endTime, 0, Conf.Limit, api.Type)
	log.Debugf("response: %v", response)
	if response != nil && response.NumOfRecords > 0 {
		nextRecord := writeResponse(response, user, apiURL)
		for nextRecord > 0 {
			response = callAPI(apiURL, user, startTime, endTime, nextRecord, Conf.Limit, api.Type)
			nextRecord = writeResponse(response, user, apiURL)
		}
	}
}

//Trigger the APIs periodically at specified interval and writes response to responseDest directory.
func trigger(ticker *time.Ticker, baseURL string, api APIConf, user *User) {
	for {
		<-ticker.C
		call(baseURL, api, user)
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
func callAPI(apiURL string, user *User, startTime string, endTime string, indx int, limit int, apiType string) *GetAPIResponse {
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return nil
	}
	//Lock the SessionToken object before calling API
	user.sessionToken.access.Lock()
	request.Header.Set(authorizationHeader, user.sessionToken.accessToken)
	user.sessionToken.access.Unlock()

	//Adding query params
	query := request.URL.Query()
	query.Add(startTimeQueryParam, startTime)
	query.Add(endTimeQueryParam, endTime)
	query.Add(limitQueryParam, strconv.Itoa(limit))
	query.Add(indexQueryParam, strconv.Itoa(indx))
	if apiType != "" {
		query.Add(alarmTypeQueryParam, apiType)
	}
	request.URL.RawQuery = query.Encode()
	log.Info(startTimeQueryParam, ": ", query[startTimeQueryParam])
	log.Info(endTimeQueryParam, ": ", query[endTimeQueryParam])
	log.Info("URL:", request.URL)

	response, err := doRequest(request)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return nil
	}

	//Map the received response to getAPIResponse struct
	resp := new(GetAPIResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to decode response")
		return nil
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return nil
	}
	log.Infof("%s called successfully for %s.", apiURL, user.Email)

	//storing LastReceivedDataTime timestamp value to file
	err = storeLastReceivedDataTime(user, apiURL, resp.Data, resp.Type)
	if err != nil {
		log.Error(err)
	}
	return resp
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
	body, err := ioutil.ReadAll(response.Body)
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
func writeResponse(response *GetAPIResponse, user *User, apiURL string) int {
	log.Infof("Writing response for %s to %s", user.Email, user.ResponseDest+"/"+path.Base(apiURL))
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
	json.Indent(&out, []byte(response.Data), "", "  ")
	log.Infof("Writing response to file %s for %s", fileName, user.Email)
	err := writeFile(fileName, out.Bytes())
	if err != nil {
		log.Error(err)
	}
	return response.NextRecord
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
func storeLastReceivedDataTime(user *User, apiURL string, data string, respType string) error {
	var pmData interface{}
	err := json.Unmarshal([]byte(data), &pmData)
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
	for _, value := range pmData.([]interface{}) {
		metricTime := value.(map[string]interface{})[sourceField].(map[string]interface{})[fieldName].(string)
		receivedData[metricTime] = struct{}{}
	}
	eventTimes := []time.Time{}
	for key := range receivedData {
		eventTime, err := time.Parse(eventTimeFormatBaiCell, key)
		if err != nil {
			eventTime, err = time.Parse(eventTimeFormatNokiaCell, key)
			if err != nil {
				return fmt.Errorf("Unable to write last received data time, error: %v", err)
			}
		}
		eventTimes = append(eventTimes, eventTime)
	}
	sort.Slice(eventTimes, func(i, j int) bool { return eventTimes[i].Before(eventTimes[j]) })

	fileName := lastReceivedDataTime + "_" + user.Email + "_" + path.Base(apiURL)
	log.Debug("Writing to ", fileName)
	err = writeFile(fileName, []byte(truncateSeconds(eventTimes[len(eventTimes)-1]).Format(time.RFC3339)))
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
func getTimeInterval(user *User, apiURL string, interval int) (string, string) {
	currentTime := currentTime()
	//calculating 15 minutes time frame
	diff := currentTime.Minute() - (currentTime.Minute() / interval * interval) + interval
	begTime := currentTime.Add(time.Duration(-1*diff) * time.Minute)
	endTime := begTime.Add(time.Duration(interval) * time.Minute)

	//Reading start time value from file
	apiLastReceivedFile := lastReceivedDataTime + "_" + user.Email + "_" + path.Base(apiURL)
	data, err := ioutil.ReadFile(apiLastReceivedFile)
	if err != nil || len(data) == 0 {
		return truncateSeconds(begTime).Format(time.RFC3339), truncateSeconds(endTime).Format(time.RFC3339)
	}
	return strings.TrimSpace(string(data)), truncateSeconds(endTime).Format(time.RFC3339)
}
