/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package ndacapis

import (
	"bytes"
	"collector/utils"
	"encoding/json"
	"net/http"

	"collector/config"

	log "github.com/sirupsen/logrus"
)

//NhgDetailsResponse struct for listNhgData API.
type NhgDetailsResponse struct {
	Status     Status      `json:"status"` // Status of the response
	NhgDetails interface{} `json:"network_info"`
}

const (
	nhgStatusField = "nhg_config_status"
	nhgIDField     = "nhg_id"

	activeNhgStatus = "ACTIVE"
)

//get nhg details for the customer
func getNhgDetails(api *config.APIConf, user *config.User, txnID uint64) {
	apiURL := config.Conf.BaseURL + api.API
	if !user.IsSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api_url": apiURL}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, utils.CurrentTime())
		return
	}

	//wait if refresh token api is running
	user.Wg.Wait()

	log.WithFields(log.Fields{"tid": txnID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	response, err := doRequest(request)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	//Map the received response to NhgDetailsResponse struct
	resp := new(NhgDetailsResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return
	}
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "nhg_ids": resp.NhgDetails}).Info("Response")

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return
	}

	err = utils.WriteResponse(user, api, resp.NhgDetails, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}

	user.NhgIDs = []string{}
	for _, nhgInfo := range resp.NhgDetails.([]interface{}) {
		nhgDetails := nhgInfo.(map[string]interface{})
		if nhgDetails[nhgStatusField].(string) == activeNhgStatus {
			user.NhgIDs = append(user.NhgIDs, nhgDetails[nhgIDField].(string))
		}
	}

	if len(user.NhgIDs) == 0 {
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no active nhg found for user")
		user.IsSessionAlive = false
	} else {
		user.IsSessionAlive = true
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "nhg_ids": user.NhgIDs}).Info("user nhgs")
	}
}
