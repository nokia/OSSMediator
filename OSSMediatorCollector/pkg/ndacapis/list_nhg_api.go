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
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
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
	orgResponse, error := fetchOrgUUID(api, user, txnID)
	if error != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": error}).Errorf("Error while fetching oruuid")
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
		if error != nil {
			user.IsSessionAlive = false
			log.WithFields(log.Fields{"tid": txnID, "error": error}).Errorf("Error while fetching oruuid")
			return
		}
		for _, acc := range accResponse.AccDetails {
			fmt.Println("Hi")
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
			storeUserNhg(nhgData.NetworkInfo, user, org, acc, txnID)
			storeUserHwID(nhgData.NetworkInfo, user, org, acc, txnID)
		}
	}

}

func storeUserNhg(nhgData []NetworkInfo, user *config.User, org config.OrgDetails, acc config.AccDetails, txnID uint64) {
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails = org
	orgAcc.AccDetails = acc

	user.NhgIDs = map[string]config.OrgAccDetails{}
	for _, nhgInfo := range nhgData {
		if nhgInfo.NhgConfigStatus == activeNhgStatus {
			user.NhgIDs[nhgInfo.NhgID] = orgAcc
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

func storeUserHwID(nhgData []NetworkInfo, user *config.User, org config.OrgDetails, acc config.AccDetails, txnID uint64) {
	orgAcc := config.OrgAccDetails{}
	orgAcc.OrgDetails = org
	orgAcc.AccDetails = acc

	user.HwIDs = map[string]config.OrgAccDetails{}
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
		user.HwIDs[hwID] = orgAcc
	}
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "hw_ids": user.HwIDs}).Info("user's access point hardware")
}
