/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

//Config read from resources/conf.json file
type Config struct {
	SourceDirs       []string `json:"source_dirs"`
	ElasticsearchURL string   `json:"elasticsearch_url"`
	CleanupDuration  int      `json:"cleanup_duration"`
}

//ReadConfig reads the configurations from conf.json file
func ReadConfig(confFile string) (Config, error) {
	var config Config
	contents, err := ioutil.ReadFile(confFile)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(contents, &config)
	if err != nil {
		return config, err
	}

	//trimming whitespaces
	config.ElasticsearchURL = strings.TrimSpace(config.ElasticsearchURL)
	for i, dir := range config.SourceDirs {
		config.SourceDirs[i] = strings.TrimSpace(dir)
	}

	if !isURLValid(config.ElasticsearchURL) {
		return config, fmt.Errorf("invalid url: %s", config.ElasticsearchURL)
	}

	log.Info("Config read successfully.")
	return config, nil
}

func isURLValid(elkURL string) bool {
	_, err := url.ParseRequestURI(elkURL)
	if err != nil {
		return false
	}
	return true
}
