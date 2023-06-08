/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package config

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

type UserAGConf struct {
	ListOrgUUID string `json:"list_orgUUID"` //fetch OrgUUID API
	ListAccUUID string `json:"list_accUUID"` //fetch acc UUID API
}

// Config keeps the config from json
type Config struct {
	BaseURL          string     `json:"base_url"` //Base URL of the API
	AzureSessionAPIs AzureConf  `json:"azure_session_api"`
	UMAPIs           UMConf     `json:"um_api"`       //User management Configuration
	ListNhGAPI       *APIConf   `json:"list_nhg_api"` //list NHG API to keep track of all ACTIVE NHG of the user.
	MetricAPIs       []*APIConf `json:"metric_apis"`  //Array of API config
	SimAPIs          []*APIConf `json:"sim_apis"`     //Array of API config
	UserAGAPIs       UserAGConf `json:"userAG_apis"`  //Array of API config
	Users            []*User    `json:"users"`        //Keep track of all the user's details
	Limit            int        `json:"limit"`
	Delay            int        `json:"delay"`
}

type OrgDetails struct {
	OrgUUID  string `json:"organization_uuid"`
	OrgAlias string `json:"organization_alias"`
}

type AccDetails struct {
	AccUUID  string `json:"account_uuid"`
	AccAlias string `json:"account_alias"`
}

type OrgAccDetails struct {
	OrgDetails OrgDetails
	AccDetails AccDetails
}

// User keeps Login configurations
type User struct {
	Email          string        `json:"email_id"`      //User's email ID
	Password       string        `json:"password"`      //User's password read from configuration file
	AuthType       string        `json:"auth_type"`     //authentication type
	ResponseDest   string        `json:"response_dest"` //Base directory where sub-directories will be created for each APIs to store its response.
	SessionToken   *SessionToken //SessionToken variable keeps track of access_token, refresh_token and expiry_time of the token. It is used for authenticating the API calls.
	Wg             sync.WaitGroup
	IsSessionAlive bool
	NhgIDsABAC     map[string]OrgAccDetails
	HwIDsABAC      map[string]OrgAccDetails
	NhgIDs         []string
	HwIDs          []string
}

// SessionToken struct tracks the access_token, refresh_token and expiry_time of the token
// As the session token will be shared by multiple APIs.
type SessionToken struct {
	AccessToken  string
	RefreshToken string
	ExpiryTime   time.Time
}

// UMConf keeps user management APIs
type UMConf struct {
	Login   string `json:"login"`   //Login API
	Refresh string `json:"refresh"` //Refresh session URL
	Logout  string `json:"logout"`  //Logout API
}

type AzureConf struct {
	Refresh string `json:"refresh"`
}

// APIConf keeps API configs
type APIConf struct {
	API          string `json:"api"`           //API URL
	Type         string `json:"type"`          //API type, HISTORY or ACTIVE from FM API.
	MetricType   string `json:"metric_type"`   //Metrics type for the API, RADIO and DAC from FM API.
	Interval     int    `json:"interval"`      //Interval at which the API will be triggered periodically.
	SyncDuration int    `json:"sync_duration"` //Interval in minutes for which duration FM will be re-synced.
}

var (
	//Conf keeps the config from json and console
	Conf Config
)

// ReadConfig reads the configurations from resources/conf.json file and sets the Config object.
func ReadConfig(confFile string) error {
	contents, err := ioutil.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("error while reading conf file: %v", err)
	}
	err = json.Unmarshal(contents, &Conf)
	if err != nil {
		return fmt.Errorf("invalid conf file: %v", err)
	}

	//trim spaces
	Conf.BaseURL = strings.TrimSpace(Conf.BaseURL)
	Conf.AzureSessionAPIs.Refresh = strings.TrimSpace(Conf.AzureSessionAPIs.Refresh)
	Conf.UMAPIs.Login = strings.TrimSpace(Conf.UMAPIs.Login)
	Conf.UMAPIs.Logout = strings.TrimSpace(Conf.UMAPIs.Logout)
	Conf.UMAPIs.Refresh = strings.TrimSpace(Conf.UMAPIs.Refresh)

	fmt.Println("config refresh url : ", Conf.AzureSessionAPIs.Refresh)

	Conf.UserAGAPIs.ListOrgUUID = strings.TrimSpace(Conf.UserAGAPIs.ListOrgUUID)
	Conf.UserAGAPIs.ListAccUUID = strings.TrimSpace(Conf.UserAGAPIs.ListAccUUID)

	for _, api := range Conf.MetricAPIs {
		api.API = strings.TrimSpace(api.API)
		api.Type = strings.TrimSpace(api.Type)
		api.MetricType = strings.TrimSpace(api.MetricType)
	}

	for _, api := range Conf.SimAPIs {
		api.API = strings.TrimSpace(api.API)
		api.Type = strings.TrimSpace(api.Type)
		api.MetricType = strings.TrimSpace(api.MetricType)
	}

	for _, user := range Conf.Users {
		user.Email = strings.TrimSpace(user.Email)
		user.ResponseDest = strings.TrimSpace(user.ResponseDest)
		user.AuthType = strings.TrimSpace(user.AuthType)
	}

	log.Info("Config read successfully.")
	return nil
}
