/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"fmt"
	"io/ioutil"
	"opennmsplugin/config"
	"opennmsplugin/formatter"
	"os"
	"path/filepath"
	"strings"

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
				dirPath := user.SourceDir + "/" + f.Name()
				log.Infof("Adding watcher to %s", dirPath)
				err = add(dirPath)
				if err != nil {
					return fmt.Errorf("Error while adding watcher to %s, error: %v", dirPath, err)
				}
				processExistingFiles(dirPath, conf)
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
				user := getUserInfoForEvent(conf, event.Name)
				if strings.Contains(filepath.Base(event.Name), "pm") {
					formatter.FormatPMData(event.Name, user.PMConfig)
				} else if strings.Contains(filepath.Base(event.Name), "fm") {
					formatter.FormatFMData(event.Name, user.FMConfig, conf.OpenNMSAddress)
				}
			}
		case err := <-watcher.Errors:
			log.Error(err)
		}
	}
}

//finds the user's details for which file write event has occurred.
func getUserInfoForEvent(conf config.Config, eventName string) config.UserConf {
	var foundUser config.UserConf
	for _, user := range conf.UsersConf {
		if strings.HasPrefix(eventName, user.SourceDir) {
			foundUser = user
			break
		}
	}
	return foundUser
}

//Process all the existing collected file from PM/FM APIs using collector.
func processExistingFiles(directory string, conf config.Config) {
	//cleanup files from the directory which were written 60 mins ago
	RemoveFiles(60, directory)
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Error(err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		user := getUserInfoForEvent(conf, directory)
		if strings.Contains(filepath.Base(directory), "fm") {
			formatter.FormatFMData(directory+"/"+file.Name(), user.FMConfig, conf.OpenNMSAddress)
		} else if strings.Contains(filepath.Base(directory), "pm") {
			formatter.FormatPMData(directory+"/"+file.Name(), user.PMConfig)
		}
	}
}
