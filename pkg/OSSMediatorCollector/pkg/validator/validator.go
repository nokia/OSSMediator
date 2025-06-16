/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package validator

import (
	"collector/pkg/config"
	"fmt"
	"net/url"
	"strings"
)

// ValidateConf validates the parameters from json file
func ValidateConf(conf config.Config) error {
	if len(conf.Users) == 0 {
		return fmt.Errorf("number of users can't be zero")
	}

	if len(conf.MetricAPIs) == 0 {
		return fmt.Errorf("number of APIs can't be zero")
	}

	if !isURLValid(conf.BaseURL) {
		return fmt.Errorf("invalid url: %s", conf.BaseURL)
	}

	if conf.ListNetworkAPI == nil {
		return fmt.Errorf("list_network_api can't be empty")
	}

	if conf.ListNetworkAPI.NhgAPI == "" {
		return fmt.Errorf("list_network_api API URL can't be empty")
	}

	if conf.ListNetworkAPI.Interval == 0 {
		return fmt.Errorf("list_network_api API call interval can't be zero")
	}

	for _, api := range conf.MetricAPIs {
		if api.API == "" {
			return fmt.Errorf("API URL can't be empty")
		}
		if api.Interval == 0 {
			return fmt.Errorf("API call interval can't be zero")
		}
		if strings.Contains(api.API, "pm") && api.Type != "" {
			return fmt.Errorf("API type for pmdata should be empty")
		}
		if strings.Contains(api.API, "pm") && !(api.MetricType == "RADIO" || api.MetricType == "EDGE" || api.MetricType == "CORE" || api.MetricType == "APPLICATION" || api.MetricType == "IXR") {
			return fmt.Errorf("API metric type for pmdata should be RADIO/CORE/EDGE/APPLICATION/IXR")
		}
		if strings.Contains(api.API, "fm") && !(api.Type == "HISTORY" || api.Type == "ACTIVE") {
			return fmt.Errorf("API type for fmdata should be HISTORY/ACTIVE")
		}
		if strings.Contains(api.API, "fm") && !(api.MetricType == "RADIO" || api.MetricType == "DAC" || api.MetricType == "CORE" || api.MetricType == "APPLICATION" || api.MetricType == "IXR") {
			return fmt.Errorf("API metric type for fmdata should be RADIO/DAC/CORE/APPLICATION/IXR")
		}
	}

	if conf.UMAPIs.Login == "" {
		return fmt.Errorf("UM's login URL can't be empty")
	}

	if conf.UMAPIs.Logout == "" {
		return fmt.Errorf("UM's logout URL can't be empty")
	}

	if conf.UMAPIs.Refresh == "" {
		return fmt.Errorf("UM's refresh URL can't be empty")
	}

	if conf.Delay < 1 || conf.Delay > 15 {
		return fmt.Errorf("API delay limit should be within 1-15, delay: %d", conf.Delay)
	}

	if conf.Limit == 0 || conf.Limit > 10000 {
		return fmt.Errorf("API response limit should be within 1-10000, limit: %d", conf.Limit)
	}

	return nil
}

func isURLValid(baseURL string) bool {
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return false
	}
	return true
}
