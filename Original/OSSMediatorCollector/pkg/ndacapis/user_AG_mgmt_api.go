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

func fetchOrgUUID(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) (OrgUUIDResponse, error) {
	orgResp := OrgUUIDResponse{}
	apiURL := config.Conf.BaseURL + config.Conf.UserAGAPIs.ListOrgUUID
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "api_type": api.Type, "metric_type": api.MetricType}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())

	request, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader("{}"))
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

	if len(orgResp.OrgDetails) == 0 {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("No org_id mapped: %s for %s", apiURL, user.Email)
		return orgResp, nil
	}

	err = utils.WriteResponse(user, &config.APIConf{API: config.Conf.UserAGAPIs.ListOrgUUID}, orgResp.OrgDetails, "", txnID, prettyResponse)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}
	return orgResp, nil
}

func fetchAccUUID(api *config.APIConf, user *config.User, org config.OrgDetails, txnID uint64, prettyResponse bool) (AccUUIDResponse, error) {
	accResp := AccUUIDResponse{}
	apiURL := config.Conf.BaseURL + config.Conf.UserAGAPIs.ListAccUUID
	apiURL = strings.Replace(apiURL, "{org_uuid}", org.OrgUUID, -1)
	log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "api_type": api.Type, "metric_type": api.MetricType}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())

	request, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader("{}"))
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
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

	if len(accResp.AccDetails) == 0 {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("No account_uuid mapped: %s for %s", apiURL, user.Email)
		return accResp, nil
	}

	err = utils.WriteResponse(user, &config.APIConf{API: config.Conf.UserAGAPIs.ListAccUUID}, accResp.AccDetails, org.OrgUUID, txnID, prettyResponse)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
	}
	return accResp, nil
}
