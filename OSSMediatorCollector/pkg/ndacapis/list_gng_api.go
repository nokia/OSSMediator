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

type gngAPIResponse struct {
	Status  Status    `json:"status"`
	GngInfo []GngInfo `json:"gng_info"`
}

type GngInfo struct {
	AdminState string `json:"admin_state"`
	GngId      string `json:"gng_id"`
}

type gngAPIAllResponse struct {
	GngInfo interface{} `json:"gng_info"`
}

func getGngDetails(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) {
	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
		listGngABAC(api, user, txnID, prettyResponse)
	} else {
		listGngRBAC(api, user, txnID, prettyResponse)
	}
}

func listGngRBAC(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) {
	apiURL := config.Conf.BaseURL + api.API
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

	resp := new(gngAPIResponse)
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

	user.NhgMux.Lock()
	storeUserGngRBAC(resp.GngInfo, user)
	user.NhgMux.Unlock()
	if len(resp.GngInfo) > 0 {
		gngData := new(gngAPIAllResponse)
		_ = json.NewDecoder(bytes.NewReader(response)).Decode(&gngData)
		err = utils.WriteResponse(user, api, gngData.GngInfo, "", txnID, prettyResponse)
		if err != nil {
			log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
		}
	} else {
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no GNG found for user")
	}
	if len(user.NhgIDs) == 0 {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no active nhg found for user")
	}
}

func storeUserGngRBAC(gngData []GngInfo, user *config.User) {
	for _, gngInfo := range gngData {
		if strings.Contains(gngInfo.AdminState, "FULLY_ACTIVATED") {
			if !containsNhg(user.NhgIDs, gngInfo.GngId) {
				user.NhgIDs = append(user.NhgIDs, gngInfo.GngId)
			}
		}
	}
}

func listGngABAC(api *config.APIConf, user *config.User, txnID uint64, prettyResponse bool) {
	//wait if refresh token api is running
	user.Wg.Wait()
	apiURL := config.Conf.BaseURL + api.API
	user.NhgMux.Lock()
	defer user.NhgMux.Unlock()
	for orgID, accIDs := range user.AccountIDsABAC {
		if len(accIDs) == 0 {
			log.WithFields(log.Fields{"tid": txnID, "user": user, "org_id": orgID}).Debug("No accounts mapped")
			continue
		}

		for _, accID := range accIDs {
			log.WithFields(log.Fields{"tid": txnID}).Infof("Triggered %s for %s at %v", apiURL, user.Email, utils.CurrentTime())
			request, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
				return
			}

			request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
			query := request.URL.Query()
			query.Add(orgIDQueryParam, orgID)
			query.Add(accIDQueryParam, accID)
			request.URL.RawQuery = query.Encode()
			response, err := doRequest(request)
			if err != nil {
				log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
				return
			}

			resp := new(gngAPIResponse)
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

			storeUserGngABAC(resp.GngInfo, user, orgID, accID)
			if len(resp.GngInfo) > 0 {
				gngData := new(gngAPIAllResponse)
				_ = json.NewDecoder(bytes.NewReader(response)).Decode(&gngData)
				err = utils.WriteResponse(user, api, gngData.GngInfo, "", txnID, prettyResponse)
				if err != nil {
					log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("unable to write response for %s", user.Email)
				}
			} else {
				log.WithFields(log.Fields{"tid": txnID, "user": user.Email, "org_id": orgID, "acc_id": accID}).Info("no GNG found for user")
			}
		}
	}
	if len(user.NhgIDsABAC) == 0 {
		//user.IsSessionAlive = false
		log.WithFields(log.Fields{"tid": txnID, "user": user.Email}).Info("no active nhg found for user")
	}
}

func storeUserGngABAC(gngData []GngInfo, user *config.User, orgID, accID string) {
	orgAcc := config.OrgAccDetails{}
	orgDet := config.OrgDetails{OrgUUID: orgID}
	accDet := config.AccDetails{AccUUID: accID}

	orgAcc.OrgDetails = orgDet
	orgAcc.AccDetails = accDet

	for _, gngInfo := range gngData {
		if strings.Contains(gngInfo.AdminState, "FULLY_ACTIVATED") {
			if _, ok := user.NhgIDsABAC[gngInfo.GngId]; !ok {
				user.NhgIDsABAC[gngInfo.GngId] = orgAcc
			}
		}
	}
}

func containsNhg(nhgList []string, gng string) bool {
	for _, nhg := range nhgList {
		if nhg == gng {
			return true
		}
	}
	return false
}
