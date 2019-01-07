/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package validator

import (
	"encoding/xml"
	"io/ioutil"
	"os"

	"opennmsplugin/config"

	log "github.com/sirupsen/logrus"
)

//ValidateXML validates the generated xml from formatter syntactically.
func ValidateXML(fileName string) error {
	xmlFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer xmlFile.Close()
	content, _ := ioutil.ReadAll(xmlFile)

	var pmData struct {
		XMLName xml.Name
	}
	err = xml.Unmarshal(content, &pmData)
	return err
}

//ValidateDestDir checks if destination directory exists, if not try to create it,
// incase of error while creating directory it will terminate the program.
func ValidateDestDir(conf config.Config) {
	for _, user := range conf.UsersConf {
		validateDirectory(user.PMConfig.DestinationDir)
		validateDirectory(user.FMConfig.DestinationDir)
	}
}

//create destination dir if not existing
func validateDirectory(destDir string) {
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error while creating %s, error: %v", destDir, err)
	}
}
