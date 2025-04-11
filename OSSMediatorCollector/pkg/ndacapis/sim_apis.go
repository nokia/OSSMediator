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
	"slices"
	"strconv"
	"strings"
)

// SimAPIResponse struct for sims api response.
type SimAPIResponse struct {
	Status        Status       `json:"status"`
	EmailID       string       `json:"email_id"`
	TotalSims     int          `json:"total_sims"`
	InUseSims     int          `json:"in_use_sims"`
	PageResponse  PageResponse `json:"page_response"`
	Subsc         interface{}  `json:"subsc"`
	RequestedSims int          `json:"requested_sims"`
}

// PageResponse struct is used to get pagination information of sims api.
type PageResponse struct {
	PageDetails PageDetails `json:"page_details"`
	TotalPages  int         `json:"total_pages"`
}

// PageDetails keeps individual page details for pagination.
type PageDetails struct {
	PageNumber int `json:"page_number"`
	PageSize   int `json:"page_size"`
}

// SimData struct for storing received response from sims api.
type SimData struct {
	EmailID       string      `json:"email_id,omitempty"`
	TotalSims     int         `json:"total_sims,omitempty"`
	InUseSims     int         `json:"in_use_sims,omitempty"`
	Subsc         interface{} `json:"subsc"`
	RequestedSims int         `json:"requested_sims,omitempty"`
}

const (
	nhgPathParam       = "{nhg_id}"
	pageNoQueryParam   = "page.page_number"
	accessPointSimsAPI = "access-point-sims"
	hwIDQueryParam     = "access_point_hw_id"
)

func fetchSimData(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) {
	if !user.IsSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api": api.API}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, utils.CurrentTime())
		return
	}
	authType := strings.ToUpper(user.AuthType)
	user.NhgMux.RLock()
	defer user.NhgMux.RUnlock()
	if strings.Contains(api.API, accessPointSimsAPI) {
		if authType == "ADTOKEN" {
			log.WithFields(log.Fields{"tid": txnID, "hw_ids": len(user.HwIDsABAC)}).Infof("starting ap_sims api")
			for hwID, orgAcc := range user.HwIDsABAC {
				callAccessPointsSimAPI(api, user, hwID, orgAcc.OrgDetails.OrgUUID, orgAcc.AccDetails.AccUUID, txnID, prettyResponse)
			}
		} else {
			for _, hwID := range user.HwIDs {
				log.WithFields(log.Fields{"tid": txnID, "hw_ids": len(user.HwIDs)}).Infof("starting ap_sims api")
				callAccessPointsSimAPI(api, user, hwID, "", "", txnID, prettyResponse)
			}
		}
		log.WithFields(log.Fields{"tid": txnID, "hw_ids": len(user.HwIDs)}).Infof("finished ap_sims api")
	} else if strings.Contains(api.API, nhgPathParam) {
		if authType == "ADTOKEN" {
			for nhgID, orgAcc := range user.NhgIDsABAC {
				if sliceID, exists := user.SliceIDs[nhgID]; exists && len(config.Conf.ListNetworkAPI.SliceIDs) != 0 {
					if !slices.Contains(config.Conf.ListNetworkAPI.SliceIDs, sliceID) {
						continue
					}
				}
				callSimAPI(api, user, nhgID, orgAcc.OrgDetails.OrgUUID, orgAcc.AccDetails.AccUUID, 1, txnID, prettyResponse)
			}
		} else {
			for _, nhgID := range user.NhgIDs {
				if sliceID, exists := user.SliceIDs[nhgID]; exists && len(config.Conf.ListNetworkAPI.SliceIDs) != 0 {
					if !slices.Contains(config.Conf.ListNetworkAPI.SliceIDs, sliceID) {
						continue
					}
				}
				callSimAPI(api, user, nhgID, "", "", 1, txnID, prettyResponse)
			}
		}
	} else {
		if authType == "ADTOKEN" {
			for orgID, accIDs := range user.AccountIDsABAC {
				for _, accID := range accIDs {
					callSimAPI(api, user, "", orgID, accID, 1, txnID, prettyResponse)
				}
			}
		} else {
			callSimAPI(api, user, "", "", "", 1, txnID, prettyResponse)
		}
	}
}

func callSimAPI(api *config.APIConf, user *config.User, nhgID string, orgUUID string, accUUID string, pageNo int, txnID uint64, prettyResponse bool) {
	apiURL := config.Conf.BaseURL + api.API
	apiURL = strings.Replace(apiURL, "{nhg_id}", nhgID, -1)

	//wait if refresh token api is running
	user.Wg.Wait()

	log.WithFields(log.Fields{"tid": txnID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
	request, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	//Adding query params
	query := request.URL.Query()
	query.Add(pageNoQueryParam, strconv.Itoa(pageNo))
	query.Add(orgIDQueryParam, orgUUID)
	query.Add(accIDQueryParam, accUUID)
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
	err = utils.WriteResponse(user, api, simData, nhgID, txnID, prettyResponse)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
		return
	}

	if resp.PageResponse.TotalPages == resp.PageResponse.PageDetails.PageNumber {
		return
	} else {
		callSimAPI(api, user, nhgID, orgUUID, accUUID, pageNo+1, txnID, prettyResponse)
	}
}

func callAccessPointsSimAPI(api *config.APIConf, user *config.User, hwID string, orgUUID string, accUUID string, txnID uint64, prettyResponse bool) {
	apiURL := config.Conf.BaseURL + api.API
	//wait if refresh token api is running
	user.Wg.Wait()

	log.WithFields(log.Fields{"tid": txnID, "hw_id": hwID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
	request, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "hw_id": hwID}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	//Adding query params
	query := request.URL.Query()
	query.Add(hwIDQueryParam, hwID)
	query.Add(orgIDQueryParam, orgUUID)
	query.Add(accIDQueryParam, accUUID)
	request.URL.RawQuery = query.Encode()

	response, err := doRequest(request)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "hw_id": hwID}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	var resp struct {
		Status             Status      `json:"status"`
		AccessPointDetails interface{} `json:"access_point_imsi_details"`
	}
	err = json.NewDecoder(bytes.NewReader(response)).Decode(&resp)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "hw_id": hwID, "api_url": apiURL}).Error("Unable to decode access point sim api response")
		return
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "hw_id": hwID, "api_url": apiURL}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return
	}

	if len(resp.AccessPointDetails.([]interface{})) == 0 {
		log.WithFields(log.Fields{"tid": txnID, "hw_id": hwID}).Errorf("no access point sims found for %s", user.Email)
		return
	}
	err = utils.WriteResponse(user, api, resp.AccessPointDetails, "", txnID, prettyResponse)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err, "hw_id": hwID}).Errorf("unable to write response for %s", user.Email)
	}
}
