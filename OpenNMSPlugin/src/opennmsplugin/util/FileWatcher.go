/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
*/

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"opennmsplugin/config"
	"opennmsplugin/formatter"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var watcher *fsnotify.Watcher

//AddWatcher adds source directory specified in config that will be watched for any file creation events.
func AddWatcher(conf config.Config) error {
	var err error
	if watcher == nil {
		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("Error while creating new watcher, error: %v", err)
		}
	}

	for _, user := range conf.UsersConf {
		if user.SourceDir == "" {
			return fmt.Errorf("Source directory path can't be empty")
		}
		if _, err := os.Stat(user.SourceDir); os.IsNotExist(err) {
			return fmt.Errorf("Source directory %s not found, error: %v", user.SourceDir, err)
		}
		files, err := ioutil.ReadDir(user.SourceDir)
		if err != nil {
			return fmt.Errorf("Error while reading source directory %s, error: %v", user.SourceDir, err)
		}

		for _, f := range files {
			if f.IsDir() {
				log.Infof("Adding watcher to %s", user.SourceDir+"/"+f.Name())
				err = add(user.SourceDir + "/" + f.Name())
				if err != nil {
					return fmt.Errorf("Error while adding watcher to %s, error: %v", user.SourceDir+"/"+f.Name(), err)
				}
			}
		}
	}
	return nil
}

//Adds a directory to be watched for any file creation events.
func add(sourceDir string) error {
	err := watcher.Add(sourceDir)
	if err != nil {
		return fmt.Errorf("Unable to add watcher to %s, error %v", sourceDir, err)
	}
	log.Infof("Added watcher to %s", sourceDir)
	return nil
}

//WatchEvents watches file creation events and formats PM/FM data.
func WatchEvents(conf config.Config) {
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Debugf("Received event: %s", event.Name)
				if strings.Contains(event.Name, "pm") && strings.HasSuffix(event.Name, ".xml") {
					for _, user := range conf.UsersConf {
						if strings.HasPrefix(event.Name, user.SourceDir) {
							formatter.FormatPMData(event.Name, user.PMConfig)
							break
						}
					}

				}
				if strings.Contains(event.Name, "fm") && strings.HasSuffix(event.Name, ".csv") {
					for _, user := range conf.UsersConf {
						if strings.HasPrefix(event.Name, user.SourceDir) {
							formatter.FormatFMData(event.Name, user.FMConfig, conf.OpenNMSAddress)
							break
						}
					}
				}
			}
		case err := <-watcher.Errors:
			log.Error(err)
		}
	}
}
