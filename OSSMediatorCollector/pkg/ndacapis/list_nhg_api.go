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
)

type nhgAPIResponse struct {
	Status      Status        `json:"status"`
	NetworkInfo []NetworkInfo `json:"network_info"`
}

type NetworkInfo struct {
	Clusters []struct {
		HwSet []struct {
			HwID string `json:"hw_id"`
		} `json:"hw_set"`
	} `json:"clusters"`
	NhgID           string `json:"nhg_id"`
	NhgConfigStatus string `json:"nhg_config_status"`
}

const (
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

	resp := new(nhgAPIResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(&resp)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return
	}

	storeUserNhg(resp.NetworkInfo, user, txnID)
	storeUserHwID(resp.NetworkInfo, user, txnID)
	err = utils.WriteResponse(user, api, resp.NetworkInfo, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}
}

func storeUserNhg(nhgData []NetworkInfo, user *config.User, txnID uint64) {
	user.NhgIDs = []string{}
	for _, nhgInfo := range nhgData {
		if nhgInfo.NhgConfigStatus == activeNhgStatus {
			user.NhgIDs = append(user.NhgIDs, nhgInfo.NhgID)
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

func storeUserHwID(nhgData []NetworkInfo, user *config.User, txnID uint64) {
	user.HwIDs = []string{}
	hwIDs := make(map[string]struct{})
	for _, nhgInfo := range nhgData {
		if nhgInfo.NhgConfigStatus != activeNhgStatus {
			continue
		}
		for _, cluster := range nhgInfo.Clusters {
			for _, hwSet := range cluster.HwSet {
				hwIDs[hwSet.HwID] = struct{}{}
			}
		}
	}

	for hwID := range hwIDs {
		user.HwIDs = append(user.HwIDs, hwID)
	}
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "hw_ids": user.HwIDs}).Info("user's access point hardware")
}
