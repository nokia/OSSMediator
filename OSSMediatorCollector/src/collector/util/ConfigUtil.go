/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
	"sync"
)

//Config keeps the config from json
type Config struct {
	BaseURL string    `json:"base_url"` //Base URL of the API
	UMAPIs  UMConf    `json:"um_api"`   //User management Configuration
	APIs    []APIConf `json:"apis"`     //Array of API config
	Users   []*User   `json:"users"`    //Keep track of all the user's details
	Limit   int       `json:"limit"`
	Delay   int       `json:"delay"`
}

//User keeps Login configurations
type User struct {
	Email          string        `json:"email_id"`      //User's email ID
	Password       string        `json:"password"`      //User's password read from configuration file
	ResponseDest   string        `json:"response_dest"` //Base directory where sub-directories will be created for each APIs to store its response.
	sessionToken   *sessionToken //SessionToken variable keeps track of access_token, refresh_token and expiry_time of the token. It is used for authenticating the API calls.
	wg             sync.WaitGroup
	isSessionAlive bool
}

//UMConf keeps user management APIs
type UMConf struct {
	Login   string `json:"login"`   //Login API
	Refresh string `json:"refresh"` //Refresh session URL
	Logout  string `json:"logout"`  //Logout API
}

//APIConf keeps API configs
type APIConf struct {
	API          string `json:"api"`           //API URL
	Type         string `json:"type"`          //For getting HISTORY or ACTIVE alarm from FM API.
	Interval     int    `json:"interval"`      //Interval at which the API will be triggered periodically.
	SyncDuration int    `json:"sync_duration"` //Interval in minutes for which duration FM will be re-synced.
}

//ReadConfig reads the configurations from resources/conf.json file and sets the Config object.
func ReadConfig(confFile string) error {
	contents, err := ioutil.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("Error while reading conf file: %v", err)
	}
	err = json.Unmarshal(contents, &Conf)
	if err != nil {
		return fmt.Errorf("Invalid conf file: %v", err)
	}

	//trim spaces
	Conf.BaseURL = strings.TrimSpace(Conf.BaseURL)
	Conf.UMAPIs.Login = strings.TrimSpace(Conf.UMAPIs.Login)
	Conf.UMAPIs.Logout = strings.TrimSpace(Conf.UMAPIs.Logout)
	Conf.UMAPIs.Refresh = strings.TrimSpace(Conf.UMAPIs.Refresh)

	for _, api := range Conf.APIs {
		api.API = strings.TrimSpace(api.API)
		api.Type = strings.TrimSpace(api.Type)
	}

	for _, user := range Conf.Users {
		user.Email = strings.TrimSpace(user.Email)
		user.Password = strings.TrimSpace(user.Password)
		user.ResponseDest = strings.TrimSpace(user.ResponseDest)
		decodedPwd, err := base64.StdEncoding.DecodeString(user.Password)
		if err != nil {
			return fmt.Errorf("Unable to decode password for %v, Error: %v", user.Email, err)
		}
		user.Password = string(decodedPwd)
	}
	log.Info("Config read successfully.")
	return nil
}
