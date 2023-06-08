package ndacapis

import (
	"bytes"
	"collector/pkg/config"
	"collector/pkg/utils"
	"encoding/json"
	"fmt"
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

func fetchOrgUUID(api *config.APIConf, user *config.User, txnID uint64) (OrgUUIDResponse, error) {
	orgResp := OrgUUIDResponse{}
	apiURL := config.Conf.BaseURL + config.Conf.UserAGAPIs.ListOrgUUID
	fmt.Println("detchorguuid user auth is :", user.AuthType)
	fmt.Println("fetchorguuid azure refresh url is: ", config.Conf.AzureSessionAPIs.Refresh)
	fmt.Println("url for fetchorguuid: ", apiURL)
	fmt.Println("refresh url from fetchorguud", config.Conf.BaseURL+config.Conf.AzureSessionAPIs.Refresh)
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "api_type": api.Type, "metric_type": api.MetricType}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())

	request, err := http.NewRequest("POST", apiURL, strings.NewReader("{}"))
	if !user.IsSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api_url": apiURL}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, utils.CurrentTime())
		return orgResp, err
	}

	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return orgResp, err
	}
	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	response, err := doRequest(request)

	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return orgResp, err
	}

	err = json.NewDecoder(bytes.NewReader(response)).Decode(&orgResp)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return orgResp, err
	}

	//check response for status code
	err = checkStatusCode(orgResp.Status)
	if err != nil {
		user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return orgResp, err
	}

	err = utils.WriteResponse(user, api, orgResp.OrgDetails, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}
	return orgResp, nil
}

func fetchAccUUID(api *config.APIConf, user *config.User, org config.OrgDetails, txnID uint64) (AccUUIDResponse, error) {
	accResp := AccUUIDResponse{}
	apiURL := config.Conf.BaseURL + config.Conf.UserAGAPIs.ListAccUUID
	apiURL = strings.Replace(apiURL, "{org_uuid}", org.OrgUUID, -1)
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "api_type": api.Type, "metric_type": api.MetricType}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())

	request, err := http.NewRequest("POST", apiURL, strings.NewReader("{}"))
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return accResp, err
	}

	if !user.IsSessionAlive {
		log.WithFields(log.Fields{"tid": txnID, "api_url": apiURL}).Warnf("Skipping API call for %s at %v as user's session is inactive", user.Email, utils.CurrentTime())
		return accResp, err
	}
	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	response, err := doRequest(request)
	if err != nil || len(response) == 0 {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return accResp, err
	}

	err = json.NewDecoder(bytes.NewReader(response)).Decode(&accResp)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Error("Unable to decode response")
		return accResp, err
	}

	//check response for status code
	err = checkStatusCode(accResp.Status)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return accResp, err
	}

	err = utils.WriteResponse(user, api, accResp.AccDetails, "", txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}
	return accResp, nil
}
