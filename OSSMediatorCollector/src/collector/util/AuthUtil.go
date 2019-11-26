/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

//UMResponse keeps track of response received from UM APIs
type UMResponse struct {
	UAT struct {
		AccessToken string `json:"access_token"` //access token
	} `json:"uat"`
	RT struct {
		RefreshToken string `json:"refresh_token"` //refresh token
	} `json:"rt"`
	Status Status `json:"status"` // Status of the response
}

//LoginRequestBody to form the request body for login API.
type LoginRequestBody struct {
	EmailID  string `json:"email_id"` //User's Email ID
	Password string `json:"password"` //User's password
}

//RefreshAndLogoutRequestBody to form the request body for refresh and logout API.
type RefreshAndLogoutRequestBody struct {
	RefreshToken string `json:"refresh_token"` //refresh token
}

//Login authenticates the BaseURL with email ID and password,
// store the session token to SessionToken.
//If successful it returns nil, if there is any error it return error.
func Login(user *User) error {
	//forming the request body in following format
	//{"email_id": "string", "password": "string"}
	reqBody := LoginRequestBody{
		EmailID:  user.Email,
		Password: user.Password,
	}
	body, _ := json.Marshal(reqBody)
	request, err := http.NewRequest("POST", Conf.BaseURL+Conf.UMAPIs.Login, bytes.NewBuffer(body))
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
		return fmt.Errorf("Unable to decode response received from login API for %s, error:%v", user.Email, err)
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		return err
	}

	//Set SessionToken
	setToken(resp, user)

	log.Infof("Login successful for %s", user.Email)
	fmt.Printf("\nLogin successful for %s", user.Email)
	return nil
}

//Extracts the expiry time from access_token and set it to SessionToken.
func setToken(response *UMResponse, user *User) {
	//getting expiry time using jwt
	token, _ := jwt.Parse(response.UAT.AccessToken, nil)
	claims := token.Claims.(jwt.MapClaims)
	exp := int64(claims["exp"].(float64))
	expTime := time.Unix(exp, 0)

	user.sessionToken = &sessionToken{
		accessToken:  response.UAT.AccessToken,
		refreshToken: response.RT.RefreshToken,
		expiryTime:   expTime,
	}
	log.Debugf("Expiry time: %v for %s", user.sessionToken.expiryTime, user.Email)
}

//RefreshToken refreshes the session token before expiry_time.
//Input parameter apiUrl is the API URL for refreshing session.
func RefreshToken(user *User) {
	apiURL := Conf.BaseURL + Conf.UMAPIs.Refresh
	duration := getRefreshDuration(user)
	refreshTimer := time.NewTimer(duration)
	for {
		<-refreshTimer.C
		user.wg.Add(1)
		err := callRefreshAPI(apiURL, user)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Refresh token failed for %s, retrying to login", user.Email)
			err := Login(user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Login Failed for %s.", user.Email)
				go retryLogin(initialBackoff, user)
			} else {
				user.wg.Done()
			}
		} else {
			user.wg.Done()
		}
		user.wg.Wait()
		log.Infof("Token refreshed for %s.", user.Email)
		duration = getRefreshDuration(user)
		refreshTimer.Reset(duration)
	}
}

//Return the expiry duration.
func getRefreshDuration(user *User) time.Duration {
	duration := user.sessionToken.expiryTime.Sub(currentTime())
	duration -= 30 * time.Second
	log.Debugf("Refresh duration for %s: %v", user.Email, duration)
	return duration
}

//calls the refresh API, return nil when successful.
func callRefreshAPI(apiURL string, user *User) error {
	log.Infof("Refreshing token for %s", user.Email)
	user.sessionToken.access.Lock()
	//forming body for refresh session API
	reqBody := RefreshAndLogoutRequestBody{
		RefreshToken: user.sessionToken.refreshToken,
	}
	body, _ := json.Marshal(reqBody)
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		user.sessionToken.access.Unlock()
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authorizationHeader, user.sessionToken.accessToken)
	user.sessionToken.access.Unlock()
	response, err := doRequest(request)
	if err != nil {
		return err
	}

	//Map the received response to umResponse struct
	resp := new(UMResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		return fmt.Errorf("Unable to decode response received from refresh API for %s, error:%v", user.Email, err)
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		return err
	}
	setToken(resp, user)
	return nil
}

func retryLogin(backoff time.Duration, user *User) {
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
			user.wg.Done()
			return
		}
	}
}

//Logout to close the session.
//If successful it returns nil, if there is any error it return error.
func Logout(user *User) error {
	log.Infof("Logging out from %s for user %s.", Conf.BaseURL, user.Email)
	//forming body for logout API
	reqBody := RefreshAndLogoutRequestBody{
		RefreshToken: user.sessionToken.refreshToken,
	}
	body, _ := json.Marshal(reqBody)
	request, err := http.NewRequest("POST", Conf.BaseURL+Conf.UMAPIs.Logout, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authorizationHeader, user.sessionToken.accessToken)
	response, err := doRequest(request)
	if err != nil {
		return err
	}

	//Map the received response to umResponse struct
	resp := new(UMResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		return fmt.Errorf("Unable to decode response received from logout API for %s, error:%v", user.Email, err)
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		return err
	}
	log.Infof("%s Logged out", user.Email)
	return nil
}
