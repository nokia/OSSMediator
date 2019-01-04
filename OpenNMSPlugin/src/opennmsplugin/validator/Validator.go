/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */
 
package validator

import (
	"encoding/json"
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

//ValidateJSON validates the generated JSON from formatter syntactically.
func ValidateJSON(fileName string) error {
	var js interface{}
	xmlFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer xmlFile.Close()
	content, _ := ioutil.ReadAll(xmlFile)
	err = json.Unmarshal(content, &js)
	return err
}

//CreateDestDir checks if destination directory exists, if not try to create it,
// in case of error while creating directory it will terminate the program.
func CreateDestDir(conf config.Config) {
	for _, user := range conf.UsersConf {
		createDirectory(user.PMConfig.DestinationDir)
		createDirectory(user.FMConfig.DestinationDir)
	}
}

//create destination dir if not existing
func createDirectory(destDir string) {
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error while creating %s, error: %v", destDir, err)
	}
}
