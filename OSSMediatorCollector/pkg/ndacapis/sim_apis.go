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
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

//SimAPIResponse struct for sims api response.
type SimAPIResponse struct {
	Status        Status       `json:"status"`
	EmailID       string       `json:"email_id"`
	TotalSims     int          `json:"total_sims"`
	InUseSims     int          `json:"in_use_sims"`
	PageResponse  PageResponse `json:"page_response"`
	Subsc         interface{}  `json:"subsc"`
	RequestedSims int          `json:"requested_sims"`
}

//PageResponse struct is used to get pagination information of sims api.
type PageResponse struct {
	PageDetails PageDetails `json:"page_details"`
	TotalPages  int         `json:"total_pages"`
}

//PageDetails keeps individual page details for pagination.
type PageDetails struct {
	PageNumber int `json:"page_number"`
	PageSize   int `json:"page_size"`
}

//SimData struct for storing received response from sims api.
type SimData struct {
	EmailID       string      `json:"email_id,omitempty"`
	TotalSims     int         `json:"total_sims,omitempty"`
	InUseSims     int         `json:"in_use_sims,omitempty"`
	Subsc         interface{} `json:"subsc"`
	RequestedSims int         `json:"requested_sims,omitempty"`
}

const (
	nhgPathParam     = "{nhg_id}"
	pageNoQueryParam = "page.page_number"
)

func fetchSimData(api *config.APIConf, user *config.User, txnID uint64) {
	if !user.IsSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api": api.API}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, utils.CurrentTime())
		return
	}
	if strings.Contains(api.API, nhgPathParam) {
		for _, nhgID := range user.NhgIDs {
			callSimAPI(api, user, nhgID, 1, txnID)
		}
	} else {
		callSimAPI(api, user, "", 1, txnID)
	}
}

func callSimAPI(api *config.APIConf, user *config.User, nhgID string, pageNo int, txnID uint64) {
	apiURL := config.Conf.BaseURL + api.API
	apiURL = strings.Replace(apiURL, "{nhg_id}", nhgID, -1)

	//wait if refresh token api is running
	user.Wg.Wait()

	log.WithFields(log.Fields{"tid": txnID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	//Adding query params
	query := request.URL.Query()
	query.Add(pageNoQueryParam, strconv.Itoa(pageNo))
	request.URL.RawQuery = query.Encode()

	response, err := doRequest(request)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	resp := new(SimAPIResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": nhgID, "api_url": apiURL}).Error("Unable to decode sim api response")
		return
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "nhg_id": nhgID, "api_url": apiURL}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return
	}

	simData := SimData{
		EmailID:       resp.EmailID,
		TotalSims:     resp.TotalSims,
		InUseSims:     resp.InUseSims,
		Subsc:         resp.Subsc,
		RequestedSims: resp.RequestedSims,
	}
	err = utils.WriteResponse(user, api, simData, nhgID, txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
		return
	}

	if resp.PageResponse.TotalPages == resp.PageResponse.PageDetails.PageNumber {
		return
	} else {
		callSimAPI(api, user, nhgID, pageNo+1, txnID)
	}
}
