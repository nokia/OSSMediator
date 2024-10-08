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
	Status      Status        `json:"status"`
	NetworkInfo []NetworkInfo `json:"network_info"`
}

type nhgAPIAllResponse struct {
	NhgDetails interface{} `json:"network_info"`
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
func getNhgDetails(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) {
	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
		listNhgABAC(api, user, txnID, prettyResponse)
	} else {
		listNhgRBAC(api, user, txnID, prettyResponse)
	}
}

func listNhgRBAC(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) {
	apiURL := config.Conf.BaseURL + api.API
	//wait if refresh token api is running
	user.Wg.Wait()

	log.WithFields(log.Fields{"tid": txnID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	response, err := doRequest(request)
	if err != nil {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return
	}

	resp := new(nhgAPIResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(&resp)
	if err != nil {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return
	}

	user.NhgMux.Lock()
	storeUserNhgRBAC(resp.NetworkInfo, user)
	storeUserHwIDRBAC(resp.NetworkInfo, user, txnID)
	user.NhgMux.Unlock()
	nhgData := new(nhgAPIAllResponse)
	_ = json.NewDecoder(bytes.NewReader(response)).Decode(&nhgData)
	err = utils.WriteResponse(user, api, nhgData.NhgDetails, "", txnID, prettyResponse)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}
	if len(user.NhgIDs) == 0 {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no active nhg found for user")
	}
}

func storeUserNhgRBAC(nhgData []NetworkInfo, user *config.User) {
	user.NhgIDs = []string{}
	for _, nhgInfo := range nhgData {
		if nhgInfo.NhgConfigStatus == activeNhgStatus {
			user.NhgIDs = append(user.NhgIDs, nhgInfo.NhgID)
		}
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
	//log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "hw_ids": user.HwIDs}).Debug("user's access point hardware")
}

func listNhgABAC(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) {
	//wait if refresh token api is running
	user.Wg.Wait()

	orgResponse, err := fetchOrgUUID(api, user, txnID, prettyResponse)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while fetching orguuid")
		return
	}
	if len(orgResponse.OrgDetails) == 0 {
		user.IsSessionAlive = false
		return
	}

	user.NhgMux.Lock()
	defer user.NhgMux.Unlock()
	user.HwIDsABAC = map[string]config.OrgAccDetails{}
	user.NhgIDsABAC = map[string]config.OrgAccDetails{}
	user.AccountIDsABAC = map[string][]string{}
	for _, org := range orgResponse.OrgDetails {
		accResponse, err := fetchAccUUID(api, user, org, txnID, prettyResponse)
		if err != nil {
			log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "org_id": org, "error": err}).Errorf("Error while getting account_id")
			continue
		}
		if len(accResponse.AccDetails) == 0 {
			log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "org_id": org}).Debug("No accounts mapped")
			continue
		}

		var accIDs []string
		for _, acc := range accResponse.AccDetails {
			accIDs = append(accIDs, acc.AccUUID)
		}
		user.AccountIDsABAC[org.OrgUUID] = accIDs

		for _, acc := range accResponse.AccDetails {
			apiURL := config.Conf.BaseURL + api.API
			log.WithFields(log.Fields{"tid": txnID, "org_id": org, "acc_id": acc}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
			request, err := http.NewRequest(http.MethodGet, apiURL, nil)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err, "org_id": org, "acc_id": acc}).Errorf("Error while calling %s for %s", apiURL, user.Email)
				continue
			}

			request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
			query := request.URL.Query()
			query.Add(orgIDQueryParam, org.OrgUUID)
			query.Add(accIDQueryParam, acc.AccUUID)
			request.URL.RawQuery = query.Encode()
			response, err := doRequest(request)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err, "org_id": org, "acc_id": acc}).Errorf("Error while calling %s for %s", apiURL, user.Email)
				continue
			}

			resp := new(nhgAPIResponse)
			err = json.NewDecoder(bytes.NewReader(response)).Decode(&resp)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err, "org_id": org, "acc_id": acc}).Error("Unable to decode response")
				continue
			}

			//check response for status code
			err = checkStatusCode(resp.Status)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err, "org_id": org, "acc_id": acc}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
				continue
			}

			if len(resp.NetworkInfo) == 0 {
				log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "org_id": org.OrgUUID, "acc_id": acc.AccUUID}).Info("No nhg mapped")
				continue
			}

			storeUserNhgABAC(resp.NetworkInfo, user, &org, &acc)
			storeUserHwIDABAC(resp.NetworkInfo, user, &org, &acc)

			nhgData := new(nhgAPIAllResponse)
			_ = json.NewDecoder(bytes.NewReader(response)).Decode(&nhgData)
			err = utils.WriteResponse(user, api, nhgData.NhgDetails, "", txnID, prettyResponse)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err, "org_id": org, "acc_id": acc}).Errorf("unable to write response for %s", user.Email)
			}
		}
	}
	if len(user.NhgIDsABAC) == 0 {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no active nhg found for user")
	}
}

func storeUserNhgABAC(nhgData []NetworkInfo, user *config.User, org *config.OrgDetails, acc *config.AccDetails) {
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
}

func storeUserHwIDABAC(nhgData []NetworkInfo, user *config.User, org *config.OrgDetails, acc *config.AccDetails) {
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
	//log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "hw_ids": user.HwIDsABAC}).Debug("user's access point hardware")
}
