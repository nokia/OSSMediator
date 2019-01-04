/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package config

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	UsersConf       []UserConf `json:"users_conf"`
	OpenNMSAddress  string     `json:"opennms_address"`
	CleanupDuration int        `json:"cleanup_duration"`
}

type UserConf struct {
	SourceDir string   `json:"source_dir"`
	PMConfig  PMConfig `json:"pm_config"`
	FMConfig  FMConfig `json:"fm_config"`
}

type PMConfig struct {
	ForeignID      string `json:"foreign_id"`
	DestinationDir string `json:"destination_dir"`
}

type FMConfig struct {
	Source         string `json:"source"`
	NodeID         string `json:"node_id"`
	Host           string `json:"host"`
	Service        string `json:"service"`
	DestinationDir string `json:"destination_dir"`
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
	log.Info("Config read successfully.")
	return config, nil
}
