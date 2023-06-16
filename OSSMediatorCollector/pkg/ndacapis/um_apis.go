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
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

const (
	//Backoff duration for retrying refresh token.
	initialBackoff = 60 * time.Second
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
	//{"email_id": "string", password": "string"}
	reqBody := LoginRequestBody{EmailID: user.Email,
		Password: user.Password,
	}
	body, _ := json.Marshal(reqBody)
	apiURL := config.Conf.BaseURL + config.Conf.UMAPIs.Login
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
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

	fmt.Println("Access Token expiration is :", expTime)

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
	apiURL := config.Conf.BaseURL
	authType := strings.ToUpper(user.AuthType)

	if authType == "ADTOKEN" {
		apiURL = apiURL + config.Conf.AzureSessionAPIs.Refresh
	} else {
		apiURL = apiURL + config.Conf.UMAPIs.Refresh
	}
	duration := getRefreshDuration(user)
	refreshTimer := time.NewTimer(duration)
	for {
		<-refreshTimer.C
		user.Wg.Add(1)
		err := callRefreshAPI(apiURL, user)
		if err != nil {
			if authType == "ADTOKEN" {
				count := 1
				log.WithFields(log.Fields{"error": err}).Errorf("Refresh token failed for %s, retrying to refresh again", user.Email)
				for i := 0; i < 4; i++ {
					log.Info("Inside i loop...")
					log.Info("Calling refresh API for the " + strconv.Itoa(i+1) + "th time")
					fmt.Println("Calling refresh API for the " + strconv.Itoa(i+1) + "th time")
					err = callRefreshAPI(apiURL, user)
					time.Sleep(5 * time.Second)
					count += 1
					if err != nil {
						fmt.Println("err again...so in for loop again:", err)
					} else {
						break
					}
				}
				fmt.Println("count is : ", count)
				if count == 3 && err != nil {
					user.IsSessionAlive = false
					log.WithFields(log.Fields{"error": err}).Errorf("Refresh token failed for %s after multiple retries..Please restart OSSMediator with a new token", user.Email)
					log.Info("Terminating DA OSS Collector...")
					os.Exit(0)
				} else {
					user.IsSessionAlive = true
					user.Wg.Done()
				}
			} else {
				log.WithFields(log.Fields{"error": err}).Errorf("Refresh token failed for %s, retrying to login", user.Email)
				err = Login(user)
				if err != nil {
					log.WithFields(log.Fields{"error": err}).Errorf("Login Failed for %s.", user.Email)
					user.IsSessionAlive = false
					go retryLogin(initialBackoff, user)
				} else {
					user.IsSessionAlive = true
					user.Wg.Done()
				}
			}
		} else {
			user.IsSessionAlive = true
			user.Wg.Done()
		}
		duration = getRefreshDuration(user)
		if duration < 10*time.Second {
			log.WithFields(log.Fields{"refresh_duration": duration, "user": user.Email}).Debugf("Found less refresh duration, login will be tried for %s.", user.Email)
			duration = 5 * time.Second
			user.IsSessionAlive = false
		}
		user.Wg.Wait()
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
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
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
	respAzure := new(AzureRefreshResponse)
	authType := strings.ToUpper(user.AuthType)
	if authType == "ADTOKEN" {
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
		fmt.Println("Error inside callrefreshAPI: " + err.Error())
		return err
	}
	setToken(resp, user)
	return nil
}

func retryLogin(backoff time.Duration, user *config.User) {
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
			user.Wg.Done()
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
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
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
