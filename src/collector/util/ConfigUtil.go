/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
*/

package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

//Config keeps the config from json
type Config struct {
	BaseURL string    `json:"base_url"` //Base URL of the API
	UMAPIs  UMConf    `json:"um_api"`   //User managment Configuration
	APIs    []APIConf `json:"apis"`     //Array of API config
	Users   []*User   `json:"users"`    //Keep track of all the user's details
}

//User keeps Login configurations
type User struct {
	Email        string        `json:"email_id"` //User's email read from console
	password     string        //User's password read from console
	ResponseDest string        `json:"response_dest"` //Base directory where sub-directories will be created for each APIs to store its response.
	sessionToken *sessionToken //sessionToken variable keeps track of access_token, refresh_token and expiry_time of the token. It is used for authenticating the API calls.
	wg           sync.WaitGroup
}

//UMConf keeps user management APIs
type UMConf struct {
	Login   string `json:"login"`   //Login API
	Refresh string `json:"refresh"` //Refresh session URL
	Logout  string `json:"logout"`  //Logout API
}

//APIConf keeps API configs
type APIConf struct {
	API      string `json:"api"`      //API URL
	Interval int    `json:"interval"` //Interval at which the API will be triggered periodically.
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
	log.Info("Config read successfully.")
	return nil
}

var readPassword = terminal.ReadPassword

//ReadCredentials reads password of each user from console
func ReadCredentials() {
	for _, user := range Conf.Users {
		var password string
		fmt.Printf("Enter password for %s: ", user.Email)
		bytePassword, err := readPassword(0)
		if err != nil {
			log.Fatal("Error in reading the password: ", err)
		}
		password = string(bytePassword)
		if password == "" {
			log.Fatal("Password cannot be empty")
		}
		user.password = password
		fmt.Println()
	}
}
