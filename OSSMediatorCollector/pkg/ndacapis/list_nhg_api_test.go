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
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var (
	fetchOrgUUIDResp1 = `
	{
  		"status": {
    		"status_code": "SUCCESS",
    		"status_description": {
      			"description_code": "NOT_SPECIFIED",
      			"description": "Successfully fetched organizations mapped to the user's access groups"
    		}
  		},
  		"organization_details": [
    		{
      			"organization_uuid": "test_orguuid_1",
      			"organization_alias": "TEST_ORGUUID1"
    		},
    		{
      			"organization_uuid": "test_orguuid_2",
      			"organization_alias": "TEST_ORGUUID2"
    		}
  		]
	}`

	listHwResp = `
	{
  		"status": {
    		"status_code": "SUCCESS",
    		"status_description": {
      			"description_code": "NOT_SPECIFIED",
      			"description": "Fetch Successful for the requested user"
    		}
  		},
  		"network_info": [
    		{
      			"nhg_id": "test_nhg_1",
				"nhg_alias": "nhg_alias",
      			"nhg_config_status": "ACTIVE",
				"clusters": [
        			{
          				"cluster_id": "1",
          				"cluster_alias": "cluster1",
          				"hw_set": [
            				{
              					"hw_id": "test_hw",
              					"hw_type": "cloud",
              					"serial_no": "AA-123"
            				}
          				],
          			"cluster_status": "ACTIVE"
    				}
  				],
				"nhg_config_status": "ACTIVE"
			}
		]
	}`

	listHwRespInActive = `
	{
  		"status": {
    		"status_code": "SUCCESS",
    		"status_description": {
      			"description_code": "NOT_SPECIFIED",
      			"description": "Fetch Successful for the requested user"
    		}
  		},
  		"network_info": [
    		{
      			"nhg_id": "test_nhg_1",
				"nhg_alias": "nhg_alias",
      			"nhg_config_status": "NW_CFG_UNAVAILABLE",
				"clusters": [
        			{
          				"cluster_id": "1",
          				"cluster_alias": "cluster1",
          				"hw_set": [
            				{
              					"hw_id": "test_hw",
              					"hw_type": "cloud",
              					"serial_no": "AA-123"
            				}
          				],
          			"cluster_status": "INACTIVE"
    				}
  				],
				"nhg_config_status": "NW_CFG_UNAVAILABLE"
			}
		]
	}`

	listNhgResp = `
	{
  		"status": {
    		"status_code": "SUCCESS",
    		"status_description": {
      			"description_code": "NOT_SPECIFIED",
      			"description": "Fetch Successful for the requested user"
    		}
  		},
  		"network_info": [
    		{
      			"nhg_id": "test_nhg_1",
      			"nhg_config_status": "NW_CFG_UNAVAILABLE"
    		},
    		{
      			"nhg_id": "test_nhg_2",
      			"nhg_config_status": "ACTIVE"
			},
    		{
      			"nhg_id": "test_nhg_3",
				"nhg_config_status": "CONFIG_SUCCESS"
			},
    		{
      			"nhg_id": "test_nhg_4",
      			"nhg_config_status": "CONFIG_READY"
    		},
    		{
      			"nhg_id": "test_nhg_5",
      			"nhg_config_status": "NW_CFG_UNAVAILABLE"
    		}
  		]
	}`

	listNhgRespInactive = `
	{
  		"status": {
    		"status_code": "SUCCESS",
    		"status_description": {
      			"description_code": "NOT_SPECIFIED",
      			"description": "Fetch Successful for the requested user"
    		}
  		},
  		"network_info": [
    		{
      			"nhg_id": "test_nhg_1",
      			"nhg_config_status": "NW_CFG_UNAVAILABLE"
    		},
    		{
      			"nhg_id": "test_nhg_3",
				"nhg_config_status": "CONFIG_SUCCESS"
			},
    		{
      			"nhg_id": "test_nhg_4",
      			"nhg_config_status": "CONFIG_READY"
    		},
    		{
      			"nhg_id": "test_nhg_5",
      			"nhg_config_status": "NW_CFG_UNAVAILABLE"
    		}
  		]
	}`
)

func TestStoreNhgABAC(t *testing.T) {
	resp := new(nhgAPIResponse)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	nhgMap := make(map[string]config.OrgAccDetails)
	user.NhgIDsABAC = nhgMap
	orgUUID := config.OrgDetails{OrgUUID: "org", OrgAlias: "orgalias"}
	accUUID := config.AccDetails{AccUUID: "acc", AccAlias: "accAlias"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "ADTOKEN"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listNhgResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	err := json.Unmarshal([]byte(listNhgResp), &resp)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return
	}
	storeUserNhgABAC(resp.NetworkInfo, &user, &orgUUID, &accUUID)
	if len(user.NhgIDsABAC) != 1 {
		t.Fail()
	}
	_, exists := user.NhgIDsABAC["test_nhg_2"]
	if !exists {
		t.Fail()
	}
}

func TestStoreNhgABACInActive(t *testing.T) {
	resp := new(nhgAPIResponse)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	nhgMap := make(map[string]config.OrgAccDetails)
	user.NhgIDsABAC = nhgMap
	orgUUID := config.OrgDetails{OrgUUID: "org", OrgAlias: "orgalias"}
	accUUID := config.AccDetails{AccUUID: "acc", AccAlias: "accAlias"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "ADTOKEN"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listNhgRespInactive)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	err := json.Unmarshal([]byte(listNhgRespInactive), &resp)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return
	}
	storeUserNhgABAC(resp.NetworkInfo, &user, &orgUUID, &accUUID)
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestStoreNhgRBAC(t *testing.T) {
	resp := new(nhgAPIResponse)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}

	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "PASSWORD"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listNhgResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	err := json.Unmarshal([]byte(listNhgResp), &resp)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return
	}
	storeUserNhgRBAC(resp.NetworkInfo, &user)
	if len(user.NhgIDs) != 1 {
		t.Fail()
	}
	if user.NhgIDs[0] != "test_nhg_2" {
		t.Fail()
	}
}

func TestStoreNhgRBACInactiveUser(t *testing.T) {
	resp := new(nhgAPIResponse)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}

	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "PASSWORD"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listNhgRespInactive)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	err := json.Unmarshal([]byte(listNhgRespInactive), &resp)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return
	}
	storeUserNhgRBAC(resp.NetworkInfo, &user)
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
}

func TestStoreHWABAC(t *testing.T) {
	resp2 := new(nhgAPIResponse)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	hwMap := make(map[string]config.OrgAccDetails)
	user.HwIDsABAC = hwMap
	orgUUID := config.OrgDetails{OrgUUID: "org", OrgAlias: "orgalias"}
	accUUID := config.AccDetails{AccUUID: "acc", AccAlias: "accAlias"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "ADTOKEN"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listHwResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	err := json.Unmarshal([]byte(listHwResp), &resp2)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return
	}
	storeUserHwIDABAC(resp2.NetworkInfo, &user, &orgUUID, &accUUID)
	if len(user.HwIDsABAC) != 1 {
		t.Fail()
	}
	_, exists := user.HwIDsABAC["test_hw"]
	if !exists {
		t.Fail()
	}
}

func TestStoreHWABACInactive(t *testing.T) {
	resp2 := new(nhgAPIResponse)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	hwMap := make(map[string]config.OrgAccDetails)
	user.HwIDsABAC = hwMap
	orgUUID := config.OrgDetails{OrgUUID: "org", OrgAlias: "orgalias"}
	accUUID := config.AccDetails{AccUUID: "acc", AccAlias: "accAlias"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "ADTOKEN"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listHwRespInActive)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	err := json.Unmarshal([]byte(listHwRespInActive), &resp2)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return
	}
	storeUserHwIDABAC(resp2.NetworkInfo, &user, &orgUUID, &accUUID)
	if len(user.HwIDsABAC) != 0 {
		t.Fail()
	}
}

func TestStoreHWRBAC(t *testing.T) {
	resp2 := new(nhgAPIResponse)
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}

	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "PASSWORD"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listHwResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	err := json.Unmarshal([]byte(listHwResp), &resp2)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return
	}
	storeUserHwIDRBAC(resp2.NetworkInfo, &user, 1234)
	if len(user.HwIDs) != 1 {
		t.Fail()
	}
}

func TestGetNhgDetails(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.AuthType = "PASSWORD"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listNhgResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getNhgDetail")
	getNhgDetails(&config.APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234, true)
	if len(user.NhgIDs) != 1 {
		t.Fail()
	}
	if user.NhgIDs[0] != "test_nhg_2" {
		t.Fail()
	}
}

func TestGetNhgDetailsWithInactiveSession(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: false}
	config.Conf = config.Config{
		BaseURL: "http://localhost:8080/v1",
	}
	getNhgDetails(&config.APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestGetNhgDetailsForInvalidCase(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Token sent is empty. Invalid Token"
				}
			}
		}`)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}

	CreateHTTPClient("", true)
	getNhgDetails(&config.APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234, true)
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
}

func TestGetNhgDetailsForInvalidURL(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}

	CreateHTTPClient("", true)
	config.Conf = config.Config{
		BaseURL: ":",
	}
	getNhgDetails(&config.APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "missing protocol scheme") {
		t.Fail()
	}
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
}

func TestGetNhgDetailsWithInvalidResponse(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}

	CreateHTTPClient("", true)
	getNhgDetails(&config.APIConf{API: "/getNhgDetail", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
}
