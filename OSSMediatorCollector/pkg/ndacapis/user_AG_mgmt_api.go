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

type OrgUUIDResponse struct {
	OrgDetails []config.OrgDetails `json:"organization_details"`
	Status     Status              `json:"status"` // Status of the response
}

type AccUUIDResponse struct {
	AccDetails []config.AccDetails `json:"account_details"`
	Status     Status              `json:"status"` // Status of the response
}

func fetchOrgUUID(api *config.APIConf, user *config.User, txnID uint64) OrgUUIDResponse {
	orgResp := OrgUUIDResponse{}
	apiURL := config.Conf.BaseURL + config.Conf.UserAGAPIs.ListOrgUUID
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return orgResp
	}
	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	response, err := doRequest(request)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return orgResp
	}

	err = json.NewDecoder(bytes.NewReader(response)).Decode(&orgResp)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return orgResp
	}

	//check response for status code
	err = checkStatusCode(orgResp.Status)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return orgResp
	}

	err = utils.WriteResponse(user, api, orgResp.OrgDetails, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}

	orgUUID := new(OrgUUIDResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(orgUUID)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to extract data from response")
		return orgResp
	}
	return orgResp
}

func fetchAccUUID(api *config.APIConf, user *config.User, org config.OrgDetails, txnID uint64) AccUUIDResponse {
	accResp := AccUUIDResponse{}
	apiURL := config.Conf.BaseURL + config.Conf.UserAGAPIs.ListAccUUID
	apiURL = strings.Replace(apiURL, "{org_uuid}", org.OrgUUID, -1)
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return accResp
	}
	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	response, err := doRequest(request)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return accResp
	}

	err = json.NewDecoder(bytes.NewReader(response)).Decode(&accResp)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return accResp
	}

	//check response for status code
	err = checkStatusCode(accResp.Status)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return accResp
	}

	err = utils.WriteResponse(user, api, accResp.AccDetails, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}

	accUUID := new(AccUUIDResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(accUUID)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to extract data from response")
		return accResp
	}
	return accResp
}