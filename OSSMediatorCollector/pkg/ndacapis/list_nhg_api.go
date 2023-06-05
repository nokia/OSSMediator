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
	"strings"
)

type nhgAPIResponse struct {
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

// get nhg details for the customer
func getNhgDetails(api *config.APIConf, user *config.User, txnID uint64) {
	user.HwIDsABAC = map[string]config.OrgAccDetails{}
	user.NhgIDsABAC = map[string]config.OrgAccDetails{}
	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
		listNhgABAC(api, user, txnID)
	} else {
		listNhgRBAC(api, user, txnID)
	}
}

func listNhgRBAC(api *config.APIConf, user *config.User, txnID uint64) {
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

	//Map the received response to struct
	var resp struct {
		Status     Status      `json:"status"` // Status of the response
		NhgDetails interface{} `json:"network_info"`
	}
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

	err = utils.WriteResponse(user, api, resp.NhgDetails, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}

	nhgData := new(nhgAPIResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(nhgData)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to extract data from response")
		return
	}
	storeUserNhgRBAC(nhgData.NetworkInfo, user, txnID)
	storeUserHwIDRBAC(nhgData.NetworkInfo, user, txnID)
}

func storeUserNhgRBAC(nhgData []NetworkInfo, user *config.User, txnID uint64) {
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

func storeUserHwIDRBAC(nhgData []NetworkInfo, user *config.User, txnID uint64) {
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

func listNhgABAC(api *config.APIConf, user *config.User, txnID uint64) {
	orgResponse, error := fetchOrgUUID(api, user, txnID)
	if error != nil || (len(orgResponse.OrgDetails)) == 0 {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": error}).Errorf("Error while fetching orguuid")
		return
	}
	//check response for status code
	err := checkStatusCode(orgResponse.Status)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling")
		return
	}

	err = utils.WriteResponse(user, api, orgResponse.OrgDetails, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}

	for _, org := range orgResponse.OrgDetails {
		accResponse, error := fetchAccUUID(api, user, org, txnID)
		if error != nil || len(accResponse.AccDetails) == 0 {
			log.WithFields(log.Fields{"tid": txnID, "error": error}).Errorf("No accs mapped")
			continue
		}
		for _, acc := range accResponse.AccDetails {
			apiURL := config.Conf.BaseURL + api.API + "?user_info.org_uuid=" + org.OrgUUID + "&user_info.account_uuid=" + acc.AccUUID
			if !user.IsSessionAlive {
				log.WithFields(log.Fields{"tid": txnID, "api_url": apiURL}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, utils.CurrentTime())
				return
			}

			//wait if refresh token api is running
			user.Wg.Wait()

			log.WithFields(log.Fields{"tid": txnID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
			request, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
				return
			}

			request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
			response, err := doRequest(request)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
				return
			}

			//Map the received response to struct
			var resp struct {
				Status     Status      `json:"status"` // Status of the response
				NhgDetails interface{} `json:"network_info"`
			}
			err = json.NewDecoder(bytes.NewReader(response)).Decode(&resp)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
				return
			}

			//check response for status code
			err = checkStatusCode(resp.Status)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
				return
			}

			err = utils.WriteResponse(user, api, resp.NhgDetails, "", txnID)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
			}

			nhgData := new(nhgAPIResponse)
			err = json.NewDecoder(bytes.NewReader(response)).Decode(nhgData)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to extract data from response")
				return
			}
			storeUserNhgABAC(nhgData.NetworkInfo, user, &org, &acc, txnID)
			storeUserHwIDABAC(nhgData.NetworkInfo, user, &org, &acc, txnID)
		}
	}
	if len(user.NhgIDsABAC) == 0 {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no active nhg found for user")
	}
}

func storeUserNhgABAC(nhgData []NetworkInfo, user *config.User, org *config.OrgDetails, acc *config.AccDetails, txnID uint64) {
	orgAcc := config.OrgAccDetails{}
	orgDet := config.OrgDetails{OrgUUID: org.OrgUUID, OrgAlias: org.OrgAlias}
	accDet := config.AccDetails{AccUUID: acc.AccUUID, AccAlias: acc.AccAlias}

	orgAcc.OrgDetails = orgDet
	orgAcc.AccDetails = accDet

	for _, nhgInfo := range nhgData {
		if nhgInfo.NhgConfigStatus == activeNhgStatus {
			user.NhgIDsABAC[nhgInfo.NhgID] = orgAcc
		}
	}
	if len(user.NhgIDsABAC) == 0 {
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no active nhg found for user")
		//user.IsSessionAlive = false
	} else {
		user.IsSessionAlive = true
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "nhg_ids": user.NhgIDsABAC}).Info("user nhgs")
	}
}

func storeUserHwIDABAC(nhgData []NetworkInfo, user *config.User, org *config.OrgDetails, acc *config.AccDetails, txnID uint64) {
	orgAcc := config.OrgAccDetails{}
	orgDet := config.OrgDetails{OrgUUID: org.OrgUUID, OrgAlias: org.OrgAlias}
	accDet := config.AccDetails{AccUUID: acc.AccUUID, AccAlias: acc.AccAlias}

	orgAcc.OrgDetails = orgDet
	orgAcc.AccDetails = accDet

	hwIDsMap := make(map[string]struct{})
	for _, nhgInfo := range nhgData {
		if nhgInfo.NhgConfigStatus != activeNhgStatus {
			continue
		}
		for _, cluster := range nhgInfo.Clusters {
			for _, hwSet := range cluster.HwSet {
				hwIDsMap[hwSet.HwID] = struct{}{}
			}
		}
	}

	for hwID := range hwIDsMap {
		user.HwIDsABAC[hwID] = orgAcc
	}
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "hw_ids": user.HwIDsABAC}).Info("user's access point hardware")
}
