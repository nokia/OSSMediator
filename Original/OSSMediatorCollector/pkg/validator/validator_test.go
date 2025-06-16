/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package validator

import (
	"collector/pkg/config"
	"strings"
	"testing"
)

var (
	conf = config.Config{
		BaseURL: "https://localhost:8080/api/v2",
		UMAPIs: config.UMConf{
			Login:   "/session",
			Refresh: "/refresh",
			Logout:  "/logout",
		},
		MetricAPIs: []*config.APIConf{
			{API: "/pmdata", Interval: 15, MetricType: "RADIO"},
			{API: "/fmdata", Interval: 60, Type: "ACTIVE", MetricType: "RADIO"},
			{API: "/fmdata", Interval: 15, Type: "HISTORY", MetricType: "RADIO"},
			{API: "/fmdata", Interval: 60, Type: "ACTIVE", MetricType: "DAC"},
			{API: "/fmdata", Interval: 15, Type: "HISTORY", MetricType: "DAC"},
		},
		Users: []*config.User{
			{Email: "user1@nokia.com", Password: "dGVzdDE=", ResponseDest: "/statistics/reports/user1"},
			{Email: "user2@nokia.com", Password: "dGVzdDI=", ResponseDest: "/statistics/reports/user2"},
		},
		ListNetworkAPI: &config.ListNetworkAPIConf{
			NhgAPI:   "/network-hardware-groups",
			Interval: 60,
		},
		Limit: 200,
		Delay: 7,
	}
)

func TestValidateConf(t *testing.T) {
	err := ValidateConf(conf)
	if err != nil {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidBaseURL(t *testing.T) {
	tmp := conf.BaseURL
	conf.BaseURL = "http//localhost:8080"
	defer func() { conf.BaseURL = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "invalid url") {
		t.Error(err)
	}
}

func TestValidateConfWithZeroAPI(t *testing.T) {
	tmp := conf.MetricAPIs
	conf.MetricAPIs = []*config.APIConf{}
	defer func() { conf.MetricAPIs = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "number of APIs can't be zero") {
		t.Error(err)
	}
}

func TestValidateConfWithZeroUsers(t *testing.T) {
	tmp := conf.Users
	conf.Users = []*config.User{}
	defer func() { conf.Users = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "number of users can't be zero") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidLoginAPI(t *testing.T) {
	tmp := conf.UMAPIs.Login
	conf.UMAPIs.Login = ""
	defer func() { conf.UMAPIs.Login = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "UM's login URL can't be empty") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidLogoutAPI(t *testing.T) {
	tmp := conf.UMAPIs.Logout
	conf.UMAPIs.Logout = ""
	defer func() { conf.UMAPIs.Logout = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "UM's logout URL can't be empty") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidRefreshAPI(t *testing.T) {
	tmp := conf.UMAPIs.Refresh
	conf.UMAPIs.Refresh = ""
	defer func() { conf.UMAPIs.Refresh = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "UM's refresh URL can't be empty") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidDelay(t *testing.T) {
	tmp := conf.Delay
	conf.Delay = 0
	defer func() { conf.Delay = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "API delay limit should be within 1-15") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidLimit(t *testing.T) {
	tmp := conf.Limit
	conf.Limit = 20000
	defer func() { conf.Limit = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "API response limit should be within 1-10000") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidAPI(t *testing.T) {
	tmp := conf.MetricAPIs[0].API
	conf.MetricAPIs[0].API = ""
	defer func() { conf.MetricAPIs[0].API = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "API URL can't be empty") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidAPIInterval(t *testing.T) {
	tmp := conf.MetricAPIs[0].Interval
	conf.MetricAPIs[0].Interval = 0
	defer func() { conf.MetricAPIs[0].Interval = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "API call interval can't be zero") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidAPITypeForPM(t *testing.T) {
	tmp := conf.MetricAPIs[0].Type
	conf.MetricAPIs[0].Type = "ACTIVE"
	defer func() { conf.MetricAPIs[0].Type = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "API type for pmdata should be empty") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidAPITypeForFM(t *testing.T) {
	tmp := conf.MetricAPIs[1].Type
	conf.MetricAPIs[1].Type = ""
	defer func() { conf.MetricAPIs[1].Type = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "API type for fmdata should be HISTORY/ACTIVE") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidAPIMetricTypeForFM(t *testing.T) {
	tmp := conf.MetricAPIs[1].MetricType
	conf.MetricAPIs[1].MetricType = ""
	defer func() { conf.MetricAPIs[1].MetricType = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "API metric type for fmdata should be RADIO/DAC") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidListNhgAPIInterval(t *testing.T) {
	tmp := conf.MetricAPIs
	conf.ListNetworkAPI.Interval = 0
	defer func() { conf.MetricAPIs = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "list_network_api API call interval can't be zero") {
		t.Error(err)
	}
}

func TestValidateConfWithInvalidListNhgAPI(t *testing.T) {
	tmp := conf.MetricAPIs
	conf.ListNetworkAPI = &config.ListNetworkAPIConf{}
	defer func() { conf.MetricAPIs = tmp }()
	err := ValidateConf(conf)
	if err == nil || !strings.Contains(err.Error(), "list_network_api API URL can't be empty") {
		t.Error(err)
	}
}
