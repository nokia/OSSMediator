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
)

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
