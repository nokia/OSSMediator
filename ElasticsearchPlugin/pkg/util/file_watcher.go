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

	"elasticsearchplugin/pkg/config"
	"elasticsearchplugin/pkg/elasticsearch"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var (
	watcher *fsnotify.Watcher
)

//AddWatcher adds source directory specified in config that will be watched for any file creation events.
func AddWatcher(conf config.Config) error {
	var err error
	if watcher == nil {
		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("error while creating new watcher, error: %v", err)
		}
	}

	for _, sourceDir := range conf.SourceDirs {
		if sourceDir == "" {
			return fmt.Errorf("source directory path can't be empty")
		}
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			return fmt.Errorf("source directory %s not found, error: %v", sourceDir, err)
		}
		files, err := ioutil.ReadDir(sourceDir)
		if err != nil {
			return fmt.Errorf("error while reading source directory %s, error: %v", sourceDir, err)
		}

		for _, f := range files {
			if f.IsDir() {
				dirPath := sourceDir + "/" + f.Name()
				log.Infof("Adding watcher to %s", dirPath)
				err = add(dirPath)
				if err != nil {
					return fmt.Errorf("error while adding watcher to %s, error: %v", dirPath, err)
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
		return fmt.Errorf("unable to add watcher to %s, error %v", sourceDir, err)
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
				log.Infof("Received event: %s", event.Name)
				elasticsearch.PushData(event.Name, conf.ElasticsearchConf)
			}
		case err := <-watcher.Errors:
			log.Error(err)
		}
	}
}

//Process all the existing collected file from PM/FM APIs using collector.
func processExistingFiles(directory string, conf config.Config) {
	//cleanup files from the directory which were written 60 mins ago
	if conf.CleanupDuration > 0 {
		RemoveFiles(conf.CleanupDuration, directory)
	}
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Error(err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		//push data to elk
		elasticsearch.PushData(directory+"/"+file.Name(), conf.ElasticsearchConf)
	}
}
