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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

const (
	//Backoff duration for retrying refresh token.
	initialBackoff = 90 * time.Second
	maxBackoff     = 16 * time.Minute
	multiplier     = 2
)

// UMResponse keeps track of response received from UM APIs.
type UMResponse struct {
	UAT struct {
		AccessToken string `json:"access_token"` //access token
	} `json:"uat"`
	RT struct {
		RefreshToken string `json:"refresh_token"` //refresh token
	} `json:"rt"`
	Status Status `json:"status"` // Status of the response
}

type AzureRefreshResponse struct {
	Token struct {
		AccessToken  string `json:"access_token"`  //access token
		RefreshToken string `json:"refresh_token"` //refresh token
	} `json:"token"`

	Status Status `json:"status"` // Status of the response
}

// LoginRequestBody to form the request body for login API.
type LoginRequestBody struct {
	EmailID  string `json:"email_id"` //User's Email ID
	Password string `json:"password"` //User's password
}

// RefreshAndLogoutRequestBody to form the request body for refresh and logout API.
type RefreshAndLogoutRequestBody struct {
	RefreshToken string `json:"refresh_token"` //refresh token
}

// Login authenticates the BaseURL with email ID and password,
// store the session token to SessionToken.
// If successful it returns nil, if there is any error it return error.
func Login(user *config.User) error {
	//forming the request body in following format
	//{"email_id": "string", "password": "string"}
	reqBody := LoginRequestBody{
		EmailID:  user.Email,
		Password: user.Password,
	}
	body, _ := json.Marshal(reqBody)
	apiURL := config.Conf.BaseURL + config.Conf.UMAPIs.Login
	request, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := doRequest(request)
	if err != nil {
		return err
	}

	//Map the received response to UMResponse struct
	resp := new(UMResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		return fmt.Errorf("unable to decode response received from login API for %s, error:%v", user.Email, err)
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		return err
	}

	//Set SessionToken
	setToken(resp, user)

	log.Infof("Login successful for %s", user.Email)
	fmt.Printf("\nLogin successful for %s\n", user.Email)
	return nil
}

func TokenAuthorize(user *config.User, sessionToken string) error {
	token := strings.Split(sessionToken, "\n")
	resp := new(UMResponse)
	resp.UAT.AccessToken = token[0]
	resp.RT.RefreshToken = token[1]

	setToken(resp, user)
	return nil
}

// Extracts the expiry time from access_token and set it to SessionToken.
func setToken(response *UMResponse, user *config.User) {
	//getting expiry time using jwt
	token, _ := jwt.Parse(response.UAT.AccessToken, nil)
	claims := token.Claims.(jwt.MapClaims)
	exp := int64(claims["exp"].(float64))
	expTime := time.Unix(exp, 0)

	user.SessionToken = &config.SessionToken{
		AccessToken:  response.UAT.AccessToken,
		RefreshToken: response.RT.RefreshToken,
		ExpiryTime:   expTime,
	}
	user.IsSessionAlive = true
	log.Debugf("Expiry time: %v for %s", user.SessionToken.ExpiryTime, user.Email)
}

// RefreshToken refreshes the session token before expiry_time.
// Input parameter apiUrl is the API URL for refreshing session.
func RefreshToken(user *config.User) {
	user.Wg.Add(1)
	apiURL := config.Conf.BaseURL
	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
		apiURL = apiURL + config.Conf.AzureSessionAPIs.Refresh
	} else {
		apiURL = apiURL + config.Conf.UMAPIs.Refresh
	}
	duration := getRefreshDuration(user)
	refreshTimer := time.NewTimer(duration)
	user.Wg.Done()
	for {
		<-refreshTimer.C
		user.Wg.Add(1)
		err := callRefreshAPI(apiURL, user)
		if err != nil {
			if authType == "ADTOKEN" {
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Errorf("Refresh token failed for %s, retrying again...", user.Email)
					user.IsSessionAlive = false
					done := make(chan bool, 1)
					go retryADRefresh(apiURL, initialBackoff, user, done)
					<-done
				} else {
					user.IsSessionAlive = true
				}
			} else {
				log.WithFields(log.Fields{"error": err}).Errorf("Refresh token failed for %s, retrying to login", user.Email)
				err = Login(user)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Errorf("Login Failed for %s.", user.Email)
					user.IsSessionAlive = false
					done := make(chan bool, 1)
					go retryLogin(initialBackoff, user, done)
					<-done
				} else {
					user.IsSessionAlive = true
				}
			}
		} else {
			user.IsSessionAlive = true
		}
		duration = getRefreshDuration(user)
		if duration < 10*time.Second {
			log.WithFields(log.Fields{"refresh_duration": duration, "user": user.Email}).Debugf("Found less refresh duration, login will be tried for %s.", user.Email)
			if authType == "ADTOKEN" {
				duration = 5 * time.Second
				user.IsSessionAlive = false
			} else {
				err = Login(user)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Errorf("Login Failed for %s.", user.Email)
					user.IsSessionAlive = false
					done := make(chan bool, 1)
					go retryLogin(initialBackoff, user, done)
					<-done
				} else {
					user.IsSessionAlive = true
				}
			}
		}
		user.Wg.Done()
		log.Infof("Token refreshed for %s.", user.Email)
		refreshTimer.Reset(duration)
	}
}

// Return the expiry duration.
func getRefreshDuration(user *config.User) time.Duration {
	duration := user.SessionToken.ExpiryTime.Sub(utils.CurrentTime())
	duration -= 30 * time.Second
	log.Debugf("Refresh duration for %s: %v", user.Email, duration)
	return duration
}

// calls the refresh API, return nil when successful.
func callRefreshAPI(apiURL string, user *config.User) error {
	log.Infof("Refreshing token for %s", user.Email)
	//forming body for refresh session API
	reqBody := RefreshAndLogoutRequestBody{
		RefreshToken: user.SessionToken.RefreshToken,
	}
	body, _ := json.Marshal(reqBody)
	request, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	response, err := doRequest(request)
	if err != nil {
		return err
	}

	//Map the received response to umResponse struct
	resp := new(UMResponse)
	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
		respAzure := new(AzureRefreshResponse)
		err = json.NewDecoder(bytes.NewReader(response)).Decode(respAzure)
		resp.UAT.AccessToken = respAzure.Token.AccessToken
		resp.RT.RefreshToken = respAzure.Token.RefreshToken
		resp.Status = respAzure.Status
	} else {
		err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	}
	if err != nil {
		return fmt.Errorf("unable to decode response received from refresh API for %s, error:%v", user.Email, err)
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		return err
	}
	setToken(resp, user)
	if authType == "ADTOKEN" {
		fileName := ".secret/." + user.Email
		encodedPassword := base64.StdEncoding.EncodeToString([]byte(resp.UAT.AccessToken)) + "\n" + base64.StdEncoding.EncodeToString([]byte(resp.RT.RefreshToken))
		err = os.WriteFile(fileName, []byte(encodedPassword), 0600)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Unable to store session token for %v to %v", user, fileName)
		}
	}
	return nil
}

func retryLogin(backoff time.Duration, user *config.User, done chan bool) {
	timer := time.NewTimer(backoff)
	for {
		<-timer.C
		log.Infof("Retrying to login with %s", user.Email)
		err := Login(user)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Login Failed for %s, login will be retried after %v", user.Email, backoff)
			backoff = backoff * multiplier
			if backoff >= maxBackoff {
				backoff = initialBackoff
			}
			timer.Reset(backoff)
		} else {
			user.IsSessionAlive = true
			done <- true
			return
		}
	}
}

func retryADRefresh(apiURL string, backoff time.Duration, user *config.User, done chan bool) {
	timer := time.NewTimer(backoff)
	for {
		<-timer.C
		log.Infof("Retrying to refresh for %s", user.Email)
		err := callRefreshAPI(apiURL, user)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Token refresh failed for %s, token refresh will be retried after %v", user.Email, backoff)
			backoff = backoff * multiplier
			if backoff >= maxBackoff {
				backoff = initialBackoff
			}
			timer.Reset(backoff)
		} else {
			user.IsSessionAlive = true
			done <- true
			return
		}
	}
}

// Logout to close the session.
// If successful it returns nil, if there is any error it return error.
func Logout(user *config.User) error {
	log.Infof("Logging out from %s for user %s.", config.Conf.BaseURL, user.Email)
	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
		user.IsSessionAlive = false
		log.Infof("%s Logged out", user.Email)
		return nil
	}
	//forming body for logout API
	reqBody := RefreshAndLogoutRequestBody{
		RefreshToken: user.SessionToken.RefreshToken,
	}
	body, _ := json.Marshal(reqBody)
	apiURL := config.Conf.BaseURL + config.Conf.UMAPIs.Logout
	request, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authorizationHeader, user.SessionToken.AccessToken)
	_, err = doRequest(request)
	if err != nil {
		return err
	}

	log.Infof("%s Logged out", user.Email)
	return nil
}
