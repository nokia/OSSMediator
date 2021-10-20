/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

//Config read from resources/conf.json file
type Config struct {
	SourceDirs        []string          `json:"source_dirs"`
	CleanupDuration   int               `json:"cleanup_duration"`
	ElasticsearchConf ElasticsearchConf `json:"elasticsearch"`
}

type ElasticsearchConf struct {
	URL                   string `json:"url"`
	User                  string `json:"user"`
	Password              string `json:"password"`
	DataRetentionDuration int    `json:"data_retention_duration"`
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
	config.ElasticsearchConf.URL = strings.TrimSpace(config.ElasticsearchConf.URL)
	config.ElasticsearchConf.User = strings.TrimSpace(config.ElasticsearchConf.User)
	config.ElasticsearchConf.Password = strings.TrimSpace(config.ElasticsearchConf.Password)
	for i, dir := range config.SourceDirs {
		config.SourceDirs[i] = strings.TrimSpace(dir)
	}

	if !isURLValid(config.ElasticsearchConf.URL) {
		return config, fmt.Errorf("invalid url: %s", config.ElasticsearchConf.URL)
	}
	if config.ElasticsearchConf.Password != "" {
		decodedPwd, err := base64.StdEncoding.DecodeString(config.ElasticsearchConf.Password)
		if err != nil {
			return config, fmt.Errorf("Unable to decode password for %v, Error: %v", config.ElasticsearchConf.Password, err)
		}
		config.ElasticsearchConf.Password = string(decodedPwd)
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
