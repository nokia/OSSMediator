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
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var (
	fetchOrgUUIDResp = `
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
      			"oranization_alias": "TEST_ORGUUID1"
    		},
    		{
      			"organization_uuid": "test_orguuid_2",
      			"organization_alias": "TEST_ORGUUID2"
    		}
  		]
	}`
)

func TestFetchOrgUUID(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fetchOrgUUIDResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getOrgDetail")
	_, err := fetchOrgUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, 1234, true)
	if err != nil {
		t.Fail()
	}
}

func TestFetchOrgUUIDWithInactiveSession(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: false}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.IsSessionAlive = false
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fetchOrgUUIDResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)

	_, err := fetchOrgUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, 1234, true)
	if err != nil || !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestFetchOrgUUIDForInvalidCase(t *testing.T) {
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
	fetchOrgUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, 1234, true)
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestFetchOrgUUIDForInvalidURL(t *testing.T) {
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
	fetchOrgUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "missing protocol scheme") {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestFetchOrgUUIDWithInvalidResponse(t *testing.T) {
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
	fetchOrgUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestFetchAccUUID(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}

	orgDet := config.OrgDetails{OrgUUID: "org_uuid_1", OrgAlias: "org_alias_1"}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fetchOrgUUIDResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getAccDetail")
	_, err := fetchAccUUID(&config.APIConf{API: "/getAccDetail", Interval: 15}, &user, orgDet, 1234, true)
	if err != nil {
		t.Fail()
	}
}

func TestFetchAccUUIDWithInactiveSession(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: false}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.IsSessionAlive = false
	orgDet := config.OrgDetails{OrgUUID: "org_uuid_1", OrgAlias: "org_alias_1"}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fetchOrgUUIDResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)

	_, err := fetchAccUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, orgDet, 1234, true)
	if err != nil || !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
}

func TestFetchAccUUIDForInvalidCase(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}

	orgDet := config.OrgDetails{OrgUUID: "org_uuid_1", OrgAlias: "org_alias_1"}

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
	fetchAccUUID(&config.APIConf{API: "/getAccDetail", Interval: 15}, &user, orgDet, 1234, true)
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestFetchAccUUIDForInvalidURL(t *testing.T) {
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
	fetchOrgUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "missing protocol scheme") {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestFetchAccUUIDWithInvalidResponse(t *testing.T) {
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
	orgDet := config.OrgDetails{OrgUUID: "org_uuid_1", OrgAlias: "org_alias_1"}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}

	CreateHTTPClient("", true)
	fetchAccUUID(&config.APIConf{API: "/getOrgDetail", Interval: 15}, &user, orgDet, 1234, true)
	if !strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}
