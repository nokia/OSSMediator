/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// CleanUp removes old response files from path periodically specified by cleanupDuration.
func CleanUp(cleanupDuration int, path string) {
	ticker := time.NewTicker(time.Duration(cleanupDuration) * time.Minute)
	go func() {
		for t := range ticker.C {
			log.Info("Triggered cleanup of ", path, " at ", t)
			RemoveFiles(cleanupDuration, path)
		}
	}()
}

// RemoveFiles removes old files from directory.
func RemoveFiles(cleanupDuration int, path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to read directory %v", path)
		return
	}

	for _, file := range files {
		cleanupTime := time.Now().Add(time.Duration(-1*cleanupDuration) * time.Minute)
		fileInfo, err := file.Info()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Unable to get file info")
			continue
		}
		if fileInfo.ModTime().Before(cleanupTime) {
			fileName := path + "/" + file.Name()
			log.Info("Deleted ", fileName, ", created on ", fileInfo.ModTime())
			err = os.Remove(fileName)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Error while deleting file ", fileName)
			}
		}
	}
}
