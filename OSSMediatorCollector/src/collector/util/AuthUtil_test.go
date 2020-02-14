/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

func TestRefresh(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			},
			"uat": {
				"access_token": "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJiUjRLTV8wLU1NVmxWWVBuTDV6N2hBWHNuS1h3UEN5TWx5U1J2RmRoUGNjIn0.eyJqdGkiOiIwMzI4ZDE5MS0wZDJiLTRmNTAtYWQzYy1jZmJlZTFjOTJjOWMiLCJleHAiOjE1MjA0ODU4MTYsIm5iZiI6MCwiaWF0IjoxNTIwNDgyMjE2LCJpc3MiOiJodHRwOi8va2V5Y2xvYWs6ODA4MC9hdXRoL3JlYWxtcy9EQWFhUy1EZW1vIiwiYXVkIjoiRGVtb0NsaWVudFRlc3QiLCJzdWIiOiJmOjI4MGJmZGZhLTcyMGMtNGM2Yy04NjM3LTRhOWJiZmIyZmVkNzp0ZXN0b3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJEZW1vQ2xpZW50VGVzdCIsImF1dGhfdGltZSI6MCwic2Vzc2lvbl9zdGF0ZSI6ImIzMjY1ODU5LTc5NmUtNDEwMi1iYjJhLTE5ZGIxMDFiZmIzZiIsImFjciI6IjEiLCJhbGxvd2VkLW9yaWdpbnMiOltdLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsiTkRBQy1QTk0tT1dORVIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJlbWFpbCI6InRlc3Rvd25lckBub2tpYS5jb20ifQ.VO7g1KnfyaCXzTnTMdgULlYDrH1Z8XMB18cTzMmcUd5tAn6YJjMWYrzk7WgTcJN1jcer_qiwGsd6Q_ERe4kwSCiayrVLEbkh2s5sbY5GDu0C5Rvk2HEB_SJgivMspfMe64ULnFFaPtmkuPR2AZUPWrrLh-iC7RWIfg0r8IpdqnsPSxP6ZDzsk1eDi_-6johvZa1YJAs3tTNxHa8fjJ7pb2i5gOe2hx6pkvudhoSAGVXILA2cQP6D5mYvim1pJ78r-WQIi93Aa7EVhGTBRMgqxoPMWmiU8YVc7JtDqLQ-s92ewvaU0u9-MBcHzWz26NOBvO9-wNQOj8Ngc7GqySCfAQ"
			},
			"rt": {
				"refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJiUjRLTV8wLU1NVmxWWVBuTDV6N2hBWHNuS1h3UEN5TWx5U1J2RmRoUGNjIn0.eyJqdGkiOiJiMWI1MWU0ZC0yZjVjLTRiMmItYmJhNS05NTJhOGE0MzNhZGUiLCJleHAiOjE1MjA0ODk0MTYsIm5iZiI6MCwiaWF0IjoxNTIwNDgyMjE2LCJpc3MiOiJodHRwOi8va2V5Y2xvYWs6ODA4MC9hdXRoL3JlYWxtcy9EQWFhUy1EZW1vIiwiYXVkIjoiRGVtb0NsaWVudFRlc3QiLCJzdWIiOiJmOjI4MGJmZGZhLTcyMGMtNGM2Yy04NjM3LTRhOWJiZmIyZmVkNzp0ZXN0b3duZXIiLCJ0eXAiOiJSZWZyZXNoIiwiYXpwIjoiRGVtb0NsaWVudFRlc3QiLCJhdXRoX3RpbWUiOjAsInNlc3Npb25fc3RhdGUiOiJiMzI2NTg1OS03OTZlLTQxMDItYmIyYS0xOWRiMTAxYmZiM2YiLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsiTkRBQy1QTk0tT1dORVIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19fQ.N6_bDM-dVrzY9bHKYsqoGJWjrqcUY3hrrOjvW1blZsLFvqbgWAYG_TFf3WETDOVG9Q0015P4s2ly97nT3mrr2XXboOj5SpIYProDx84Xt4UoetPFqR4ZB6VY3LemwlXxiFOHPHjbNHWDiC3pgSTRcqvzDlmWqzBNjHQljJcF9Btj466-pwPIg-ITphaXVAl_S4I2NnL7kt4H8IsoBOlcYbhW2c5oVmsny6KxOBL0ISNVfHYg_3oDUnMYpnX0s6KadIjRwLaHE0u-jQBQr8Dl9NLPehmnaLnSwY-ldaUkM9naG28iHSmPdAUqFmaPJY7k4LL1ugtztNH3D_cT_2jAtw"
			}
		}`)
	}))
	defer testServer.Close()
	var user User
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now().Add(61 * time.Second),
	}
	CreateHTTPClient("", true)
	err := callRefreshAPI(testServer.URL, &user)
	time.Sleep(10 * time.Millisecond)
	if err != nil {
		t.Error(err)
	}
}

func TestRefreshWithInvalidResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	var user User
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now().Add(61 * time.Second),
	}
	CreateHTTPClient("", true)
	err := callRefreshAPI(testServer.URL, &user)
	if err == nil || !strings.Contains(err.Error(), "Unable to decode response") {
		t.Fail()
	}
}

func TestRefreshForEmptyRefreshToken(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Refresh token sent is empty. Invalid Token"
				}
			},
			"uat": {
				"access_token": "Undefined"
			},
			"rt": {
				"refresh_token": "Undefined"
			}
		}`)
	}))
	defer testServer.Close()
	var user User
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "",
		expiryTime:   time.Now().Add(65 * time.Second),
	}
	CreateHTTPClient("", true)
	err := callRefreshAPI(testServer.URL, &user)
	if err == nil {
		t.Fail()
	}
}

func TestRefreshForWrongURL(t *testing.T) {
	var user User
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "",
		expiryTime:   time.Now().Add(65 * time.Second),
	}
	CreateHTTPClient("", true)
	err := callRefreshAPI("http://localhost:8080", &user)
	if err == nil {
		t.Fail()
	}
}

func TestRefreshForEmptyURL(t *testing.T) {
	var user User
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now().Add(65 * time.Second),
	}
	CreateHTTPClient("", true)
	err := callRefreshAPI("", &user)
	if err == nil {
		t.Fail()
	}
}

func TestRefreshForInvalidURL(t *testing.T) {
	var user User
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now().Add(65 * time.Second),
	}
	CreateHTTPClient("", true)
	err := callRefreshAPI(":", &user)
	if err == nil {
		t.Fail()
	}
}

func TestGetRefreshDuration(t *testing.T) {
	var user User
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now().Add(65 * time.Second),
	}
	duration := getRefreshDuration(&user)
	if duration < 0 {
		t.Error(duration)
	}
}

func TestLoginWithWrongUrl(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	Conf = Config{
		BaseURL: "https://localhost:8080/login",
	}
	CreateHTTPClient("", true)
	err := Login(&user)
	if err == nil {
		t.Fail()
	}
}

func TestLoginWithEmptyUrl(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	Conf = Config{
		BaseURL: "",
	}
	CreateHTTPClient("", true)
	err := Login(&user)
	if err == nil {
		t.Fail()
	}
}

func TestLoginWithInvalidUrl(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	Conf = Config{
		BaseURL: ":",
	}
	CreateHTTPClient("", true)
	err := Login(&user)
	if err == nil {
		t.Fail()
	}
}

func TestInvalidLogin(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Invalid Username"
				}
			},
			"uat": {
				"access_token": "Undefined"
			},
			"rt": {
				"refresh_token": "Undefined"
			}
		}`)
	}))
	defer testServer.Close()
	Conf = Config{
		BaseURL: testServer.URL,
		UMAPIs: UMConf{
			Login:   "/login",
			Refresh: "/login/refreshSession",
		},
	}
	CreateHTTPClient("", true)
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	err := Login(&user)
	if err == nil {
		t.Fail()
	}
}

func TestLogin(t *testing.T) {
	mySigningKey := []byte("testtoken")
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(30 * time.Second).Unix(),
		Issuer:    "test",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(mySigningKey)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			},
			"uat": {
				"access_token": "`+tokenString+`"
			},
			"rt": {
				"refresh_token": "`+tokenString+`"
			}
		}`)
	}))
	defer testServer.Close()
	Conf = Config{
		BaseURL: testServer.URL,
		UMAPIs: UMConf{
			Login:   "",
			Refresh: "/login/refreshSession",
		},
	}
	CreateHTTPClient("", true)
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	err := Login(&user)
	if err != nil {
		t.Error(err)
	}
}

func TestLoginWIthInvalidResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	Conf = Config{
		BaseURL: testServer.URL,
		UMAPIs: UMConf{
			Login:   "",
			Refresh: "/login/refreshSession",
		},
	}
	CreateHTTPClient("", true)
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	err := Login(&user)
	if err == nil || !strings.Contains(err.Error(), "Unable to decode response") {
		t.Fail()
	}
}

func TestSetToken(t *testing.T) {
	response := UMResponse{
		UAT: struct {
			AccessToken string `json:"access_token"`
		}{
			AccessToken: "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJiUjRLTV8wLU1NVmxWWVBuTDV6N2hBWHNuS1h3UEN5TWx5U1J2RmRoUGNjIn0.eyJqdGkiOiIwMzI4ZDE5MS0wZDJiLTRmNTAtYWQzYy1jZmJlZTFjOTJjOWMiLCJleHAiOjE1MjA0ODU4MTYsIm5iZiI6MCwiaWF0IjoxNTIwNDgyMjE2LCJpc3MiOiJodHRwOi8va2V5Y2xvYWs6ODA4MC9hdXRoL3JlYWxtcy9EQWFhUy1EZW1vIiwiYXVkIjoiRGVtb0NsaWVudFRlc3QiLCJzdWIiOiJmOjI4MGJmZGZhLTcyMGMtNGM2Yy04NjM3LTRhOWJiZmIyZmVkNzp0ZXN0b3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJEZW1vQ2xpZW50VGVzdCIsImF1dGhfdGltZSI6MCwic2Vzc2lvbl9zdGF0ZSI6ImIzMjY1ODU5LTc5NmUtNDEwMi1iYjJhLTE5ZGIxMDFiZmIzZiIsImFjciI6IjEiLCJhbGxvd2VkLW9yaWdpbnMiOltdLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsiTkRBQy1QTk0tT1dORVIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJlbWFpbCI6InRlc3Rvd25lckBub2tpYS5jb20ifQ.VO7g1KnfyaCXzTnTMdgULlYDrH1Z8XMB18cTzMmcUd5tAn6YJjMWYrzk7WgTcJN1jcer_qiwGsd6Q_ERe4kwSCiayrVLEbkh2s5sbY5GDu0C5Rvk2HEB_SJgivMspfMe64ULnFFaPtmkuPR2AZUPWrrLh-iC7RWIfg0r8IpdqnsPSxP6ZDzsk1eDi_-6johvZa1YJAs3tTNxHa8fjJ7pb2i5gOe2hx6pkvudhoSAGVXILA2cQP6D5mYvim1pJ78r-WQIi93Aa7EVhGTBRMgqxoPMWmiU8YVc7JtDqLQ-s92ewvaU0u9-MBcHzWz26NOBvO9-wNQOj8Ngc7GqySCfAQ",
		},
		RT: struct {
			RefreshToken string `json:"refresh_token"`
		}{
			RefreshToken: "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJiUjRLTV8wLU1NVmxWWVBuTDV6N2hBWHNuS1h3UEN5TWx5U1J2RmRoUGNjIn0.eyJqdGkiOiJiMWI1MWU0ZC0yZjVjLTRiMmItYmJhNS05NTJhOGE0MzNhZGUiLCJleHAiOjE1MjA0ODk0MTYsIm5iZiI6MCwiaWF0IjoxNTIwNDgyMjE2LCJpc3MiOiJodHRwOi8va2V5Y2xvYWs6ODA4MC9hdXRoL3JlYWxtcy9EQWFhUy1EZW1vIiwiYXVkIjoiRGVtb0NsaWVudFRlc3QiLCJzdWIiOiJmOjI4MGJmZGZhLTcyMGMtNGM2Yy04NjM3LTRhOWJiZmIyZmVkNzp0ZXN0b3duZXIiLCJ0eXAiOiJSZWZyZXNoIiwiYXpwIjoiRGVtb0NsaWVudFRlc3QiLCJhdXRoX3RpbWUiOjAsInNlc3Npb25fc3RhdGUiOiJiMzI2NTg1OS03OTZlLTQxMDItYmIyYS0xOWRiMTAxYmZiM2YiLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsiTkRBQy1QTk0tT1dORVIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19fQ.N6_bDM-dVrzY9bHKYsqoGJWjrqcUY3hrrOjvW1blZsLFvqbgWAYG_TFf3WETDOVG9Q0015P4s2ly97nT3mrr2XXboOj5SpIYProDx84Xt4UoetPFqR4ZB6VY3LemwlXxiFOHPHjbNHWDiC3pgSTRcqvzDlmWqzBNjHQljJcF9Btj466-pwPIg-ITphaXVAl_S4I2NnL7kt4H8IsoBOlcYbhW2c5oVmsny6KxOBL0ISNVfHYg_3oDUnMYpnX0s6KadIjRwLaHE0u-jQBQr8Dl9NLPehmnaLnSwY-ldaUkM9naG28iHSmPdAUqFmaPJY7k4LL1ugtztNH3D_cT_2jAtw",
		},
	}

	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	setToken(&response, &user)
	if user.sessionToken.expiryTime.String() != "2018-03-08 05:10:16 +0000 UTC" {
		t.Fail()
	}
}

func TestRefreshToken(t *testing.T) {
	mySigningKey := []byte("testtoken")
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(60 * time.Second).Unix(),
		Issuer:    "test",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(mySigningKey)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			},
			"uat": {
				"access_token": "`+tokenString+`"
			},
			"rt": {
				"refresh_token": "`+tokenString+`"
			}
		}`)
	}))
	defer testServer.Close()
	Conf = Config{
		BaseURL: testServer.URL,
		UMAPIs: UMConf{
			Login:   "/login",
			Refresh: "",
		},
	}
	CreateHTTPClient("", true)
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now().Add(30100 * time.Millisecond),
	}
	go RefreshToken(&user)
	time.Sleep(200 * time.Millisecond)
	if user.sessionToken.accessToken != tokenString && user.sessionToken.refreshToken != tokenString {
		t.Fail()
	}
}

func TestLogoutWithWrongUrl(t *testing.T) {
	Conf = Config{
		BaseURL: "https://localhost:8080",
		UMAPIs: UMConf{
			Logout: "/logout",
		},
	}
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now().Add(35 * time.Second),
	}
	CreateHTTPClient("", true)
	err := Logout(&user)
	if err == nil {
		t.Fail()
	}
}

func TestLogoutWithInvalidResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	Conf = Config{
		BaseURL: testServer.URL,
		UMAPIs: UMConf{
			Login:   "/login",
			Refresh: "/login/refreshSession",
			Logout:  "/logout",
		},
	}
	CreateHTTPClient("", true)
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now().Add(35 * time.Second),
	}
	err := Logout(&user)
	if err == nil || !strings.Contains(err.Error(), "Unable to decode response") {
		t.Fail()
	}
}

func TestLogoutWithEmptyUrl(t *testing.T) {
	Conf = Config{
		BaseURL: "",
	}
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now().Add(35 * time.Second),
	}
	CreateHTTPClient("", true)
	err := Logout(&user)
	if err == nil {
		t.Fail()
	}
}

func TestLogoutWithInvalidUrl(t *testing.T) {
	Conf = Config{
		BaseURL: ":",
	}
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now().Add(35 * time.Second),
	}
	CreateHTTPClient("", true)
	err := Logout(&user)
	if err == nil {
		t.Fail()
	}
}

func TestInvalidLogout(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Refresh token sent is empty. Invalid Token"
				}
			}
		}`)
	}))
	defer testServer.Close()
	Conf = Config{
		BaseURL: testServer.URL,
		UMAPIs: UMConf{
			Login:   "/login",
			Refresh: "/login/refreshSession",
			Logout:  "/logout",
		},
	}
	CreateHTTPClient("", true)
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now().Add(35 * time.Second),
	}
	err := Logout(&user)
	if err == nil {
		t.Fail()
	}
}

func TestLogout(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			}
		}`)
	}))
	defer testServer.Close()
	Conf = Config{
		BaseURL: testServer.URL,
		UMAPIs: UMConf{
			Login:   "/login",
			Refresh: "/login/refreshSession",
			Logout:  "/logout",
		},
	}
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now().Add(35 * time.Second),
	}
	CreateHTTPClient("", true)
	err := Logout(&user)
	if err != nil {
		t.Error(err)
	}
}

func TestRetryLogin(t *testing.T) {
	user := User{Email: "testuser@nokia.com", Password: "MTIzNA=="}
	CreateHTTPClient("", true)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	go retryLogin(100*time.Millisecond, &user)
	time.Sleep(200 * time.Millisecond)
	if !strings.Contains(buf.String(), "Retrying to login with "+user.Email) {
		t.Fail()
	}
}
