/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package ndacapis

import (
	"bytes"
	"collector/pkg/config"
	"collector/pkg/notifier"
	"collector/pkg/utils"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strconv"
	"strings"
)

// GetAPIResponse keeps track of response received from PM/FM API.
type GetAPIResponse struct {
	Type            string      `json:"type"`
	TotalNumRecords int         `json:"total_num_records"`
	NumOfRecords    int         `json:"num_of_records"`
	NextRecord      int         `json:"next_record"`
	Data            interface{} `json:"data"`
	Status          Status      `json:"status"` // Status of the response
	SearchAfterKey  string      `json:"search_after_key"`
}

// Status keeps track of status from response.
type Status struct {
	StatusCode        string            `json:"status_code"`
	StatusDescription StatusDescription `json:"status_description"`
}

// StatusDescription keeps track of status description from response.
type StatusDescription struct {
	DescriptionCode string `json:"description_code"`
	Description     string `json:"description"`
}

// ErrorResponse struct for parsing error response from APIs.
type ErrorResponse struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type apiCallRequest struct {
	url            string
	api            *config.APIConf
	user           *config.User
	nhgID          string
	startTime      string
	endTime        string
	index          int
	limit          int
	searchAfterKey string
	orgUUID        string
	accUUID        string
}

const (
	//retry msg
	retryCurrentMsg = "retry current"
	retryNextMsg    = "retry next"
)

func fetchMetricsData(api *config.APIConf, user *config.User, txnID uint64) {
	if !user.IsSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api": api.API, "api_type": api.Type, "metric_type": api.MetricType}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, utils.CurrentTime())
		return
	}
	apiKey := user.Email + "_" + path.Base(api.API) + "_" + api.MetricType
	if api.Type != "" {
		apiKey += "_" + api.Type
	}
	mux.RLock()
	_, ok := activeAPIs[apiKey]
	mux.RUnlock()
	if ok {
		log.WithFields(log.Fields{"tid": txnID, "api": api.API, "api_type": api.Type, "metric_type": api.MetricType}).Debugf("Previous API call for %s at %v is  still active", user.Email, utils.CurrentTime())
		return
	}
	mux.Lock()
	activeAPIs[apiKey] = struct{}{}
	mux.Unlock()

	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
		for nhgID, orgAcc := range user.NhgIDsABAC {
			startTime, endTime := utils.GetTimeInterval(user, api, nhgID)
			apiReq := apiCallRequest{
				api:       api,
				user:      user,
				nhgID:     nhgID,
				startTime: startTime,
				endTime:   endTime,
				index:     0,
				limit:     config.Conf.Limit,
				orgUUID:   orgAcc.OrgDetails.OrgUUID,
				accUUID:   orgAcc.AccDetails.AccUUID,
			}
			msg := callMetricAPI(apiReq, maxRetryAttempts, txnID)
			if msg == retryCurrentMsg {
				callMetricAPI(apiReq, 0, txnID)
			}
			for msg == retryNextMsg {
				apiReq.startTime, apiReq.endTime = utils.GetTimeInterval(user, api, nhgID)
				apiReq.index = 0
				log.WithFields(log.Fields{"tid": txnID}).Info("retrying next")
				msg = callMetricAPI(apiReq, maxRetryAttempts, txnID)
			}
		}
	} else {
		for _, nhgID := range user.NhgIDs {
			startTime, endTime := utils.GetTimeInterval(user, api, nhgID)
			apiReq := apiCallRequest{
				api:       api,
				user:      user,
				nhgID:     nhgID,
				startTime: startTime,
				endTime:   endTime,
				index:     0,
				limit:     config.Conf.Limit,
			}
			msg := callMetricAPI(apiReq, maxRetryAttempts, txnID)
			if msg == retryCurrentMsg {
				callMetricAPI(apiReq, 0, txnID)
			}
			for msg == retryNextMsg {
				apiReq.startTime, apiReq.endTime = utils.GetTimeInterval(user, api, nhgID)
				apiReq.index = 0
				log.WithFields(log.Fields{"tid": txnID}).Info("retrying next")
				msg = callMetricAPI(apiReq, maxRetryAttempts, txnID)
			}
		}
	}
	mux.Lock()
	delete(activeAPIs, apiKey)
	mux.Unlock()
}

func callMetricAPI(req apiCallRequest, retryAttempts int, txnID uint64) string {
	apiURL := config.Conf.BaseURL + req.api.API
	authType := strings.ToUpper(req.user.AuthType)
	if authType == "ADTOKEN" {
		apiURL = apiURL + "?user_info.org_uuid=" + req.orgUUID + "&user_info.account_uuid=" + req.accUUID
		apiURL = strings.Replace(apiURL, "{nhg_id}", req.nhgID, -1)
		req.url = apiURL
	} else {
		apiURL = strings.Replace(apiURL, "{nhg_id}", req.nhgID, -1)
		req.url = apiURL
	}
	log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("Triggered %s for %s at %v", apiURL, req.user.Email, utils.CurrentTime())
	response, err := callAPI(req, txnID)
	if err != nil {
		//retry api when 500 status code is returned
		if strings.Contains(err.Error(), "500") {
			response, err = retryAPICall(req, retryAttempts, txnID)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_url": apiURL, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("API call failed, data will be skipped...")
				return ""
			}
		} else {
			return ""
		}
	}

	if response == nil {
		log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_url": req.url, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("found nil response, resp: %v, err: %v", response, err)
		return ""
	}
	if response.NumOfRecords == 0 {
		return ""
	}
	log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "total_no_of_records": response.TotalNumRecords, "received_no_of_records": response.NumOfRecords}).Infof("Received response details")

	receivedNoOfRecords := response.NumOfRecords
	var noOfRecords int
	if response.NextRecord > 0 {
		req.index = response.NextRecord
		req.searchAfterKey = response.SearchAfterKey
		noOfRecords, err = handlePagination(req, retryAttempts, txnID)
		if err != nil {
			return retryCurrentMsg
		}
		receivedNoOfRecords += noOfRecords
		log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "total_no_of_records": response.TotalNumRecords, "received_no_of_records": receivedNoOfRecords}).Infof("Received response details after pagination")
	}

	return ""
}

func handlePagination(req apiCallRequest, retryAttempts int, txnID uint64) (int, error) {
	var receivedNoOfRecords int
	for req.index > 0 {
		response, err := callAPI(req, txnID)
		if response == nil {
			log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_url": req.url, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("found nil response, resp: %v, err: %v", response, err)
		}
		if err != nil {
			response, err = retryAPICall(req, retryAttempts, txnID)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_url": req.url, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("API call failed, will be retried from starting...")
				return 0, err
			} else {
				if response == nil {
					log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_url": req.url, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("found nil response, resp: %v, err: %v", response, err)
					return 0, nil
				}
				req.index = response.NextRecord
				req.searchAfterKey = response.SearchAfterKey
				receivedNoOfRecords += response.NumOfRecords
			}
		} else {
			if response == nil {
				log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_url": req.url, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("found nil response, resp: %v, err: %v", response, err)
				return 0, nil
			}
			req.index = response.NextRecord
			req.searchAfterKey = response.SearchAfterKey
			receivedNoOfRecords += response.NumOfRecords
		}
	}
	return receivedNoOfRecords, nil
}

// retry failed API call
func retryAPICall(req apiCallRequest, retryAttempts int, txnID uint64) (*GetAPIResponse, error) {
	var err error
	var response *GetAPIResponse
	for i := 0; i < retryAttempts; i++ {
		log.WithFields(log.Fields{"txn_id": txnID, "nhg_id": req.nhgID, "api_url": req.url, "start_time": req.startTime, "end_time": req.endTime, "limit": req.limit, "index": req.index, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Info("retrying api call")
		response, err = callAPI(req, txnID)
		if response == nil {
			log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "api_url": req.url, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Infof("err: %v, resp: %v", err, response)
		}
		if err == nil {
			return response, nil
		}
	}
	return nil, err
}

// CallAPI calls the API, adds authorization, query params and returns response.
// If successful it returns response as array of byte, if there is any error it return nil.
func callAPI(req apiCallRequest, txnID uint64) (*GetAPIResponse, error) {
	request, err := http.NewRequest("GET", req.url, nil)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", req.url, req.user.Email)
		return nil, err
	}

	//wait if refresh token api is running
	req.user.Wg.Wait()

	request.Header.Set(authorizationHeader, req.user.SessionToken.AccessToken)
	//requesting compressed response
	request.Header.Add("Accept-Encoding", "gzip")

	//Adding query params
	query := request.URL.Query()
	query.Add(startTimeQueryParam, req.startTime)
	query.Add(endTimeQueryParam, req.endTime)
	query.Add(limitQueryParam, strconv.Itoa(req.limit))
	query.Add(indexQueryParam, strconv.Itoa(req.index))
	if req.api.Type != "" {
		query.Add(alarmTypeQueryParam, req.api.Type)
	}
	if req.api.MetricType != "" {
		query.Add(metricTypeQueryParam, req.api.MetricType)
	}
	if req.searchAfterKey != "" {
		query.Add(searchAfterKeyQueryParam, req.searchAfterKey)
	}

	request.URL.RawQuery = query.Encode()
	log.WithFields(log.Fields{"tid": txnID}).Info(startTimeQueryParam, ": ", query[startTimeQueryParam])
	log.WithFields(log.Fields{"tid": txnID}).Info(endTimeQueryParam, ": ", query[endTimeQueryParam])
	log.WithFields(log.Fields{"tid": txnID}).Info("URL:", request.URL)

	response, err := doRequest(request)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": req.nhgID, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType}).Errorf("Error while calling %s for %s", req.url, req.user.Email)
		return nil, err
	}

	//Map the received response to getAPIResponse struct
	resp := new(GetAPIResponse)
	resp.Type = path.Base(req.url)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": req.nhgID, "start_time": req.startTime, "end_time": req.endTime}).Error("Unable to decode response")
		return nil, err
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": req.nhgID, "start_time": req.startTime, "end_time": req.endTime}).Errorf("Invalid status code received while calling %s for %s", req.url, req.user.Email)
		return nil, err
	}
	log.WithFields(log.Fields{"tid": txnID, "nhg_id": req.nhgID, "start_time": req.startTime, "end_time": req.endTime, "api_type": req.api.Type, "metric_type": req.api.MetricType, "total_no_of_records": resp.TotalNumRecords, "no_of_records_received": resp.NumOfRecords, "next_record_index": resp.NextRecord}).Infof("%s called successfully for %s.", req.url, req.user.Email)

	if resp.NumOfRecords == 0 {
		log.WithFields(log.Fields{"tid": txnID, "api_url": req.url, "user": req.user.Email, "total_no_of_records": resp.TotalNumRecords}).Info("no records found")
		return resp, nil
	}

	//storing LastReceivedDataTime timestamp value to file
	err = utils.StoreLastReceivedDataTime(req.user, resp.Data, req.api, req.nhgID, txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID}).Error(err)
	}

	if path.Base(req.api.API) == fmResponseType {
		go notifier.RaiseAlarmNotification(txnID, resp.Data, req.api.MetricType, req.api.Type)
	}
	//write response
	err = utils.WriteResponse(req.user, req.api, resp.Data, req.nhgID, txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", req.user.Email)
		return nil, err
	}
	return resp, nil
}
