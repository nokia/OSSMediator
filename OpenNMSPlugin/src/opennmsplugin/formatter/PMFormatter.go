/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"opennmsplugin/config"
	"opennmsplugin/validator"

	log "github.com/sirupsen/logrus"
)

const (
	eventTimeFormatBaiCell   = "2006-01-02T15:04:05Z07:00"
	eventTimeFormatNokiaCell = "2006-01-02T15:04:05Z07:00:00"
	dateFormat               = "20060102"
	timeFormat               = "1504"
	filePrefix               = "A"

	//field name to extract data from PM response file
	sourceField    = "_source"
	eventTimeField = "timestamp"
)

type response struct {
	Data interface{} `json:"data"`
}

//FormatPMData formats PM Data from A2 to A3 format
func FormatPMData(filePath string, pmConfig config.PMConfig) {
	log.Infof("Formatting PM file %s", filePath)
	f, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while formatting pm data %s", filePath)
		return
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error(err)
		return
	}
	//removing source file
	defer os.Remove(filePath)

	var pmData interface{}
	err = json.Unmarshal(content, &pmData)
	if err != nil {
		log.Error(err)
		return
	}

	for _, value := range pmData.([]interface{}) {
		metricTime := value.(map[string]interface{})[sourceField].(map[string]interface{})[eventTimeField].(string)
		resp := &response{
			Data: value,
		}
		contents, err := json.Marshal(resp)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Unable to marshal received json response in %s for %s time", filePath, metricTime)
			continue
		}
		fileName, err := renameFile(metricTime, pmConfig)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Error while formatting pm data %s for %s time", filePath, metricTime)
			continue
		}
		var out bytes.Buffer
		json.Indent(&out, contents, "", "  ")
		err = writeFile(fileName, out.Bytes())
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Error while formatting pm data %s", filePath)
			continue
		}
		//Validate the generated json file.
		err = validator.ValidateJSON(fileName)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Invalid JSON generated from %s", filePath)
			os.Remove(fileName)
			continue
		}
		log.Infof("Formatted PM data written to %s", fileName)
	}
}

//writes data to file
func writeFile(fileName string, data []byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("File creation failed: %v", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Error while writting response to file: %v", err)
	}
	return nil
}

//Rename the file to OpenNMS format
func renameFile(metricTime string, pmConfig config.PMConfig) (string, error) {
	t, err := time.Parse(eventTimeFormatBaiCell, metricTime)
	if err != nil {
		t, err = time.Parse(eventTimeFormatNokiaCell, metricTime)
		if err != nil {
			return "", err
		}
	}
	_, offSet := t.Zone()
	var offset string
	if offSet < 0 {
		offset = fmt.Sprintf("%02d", (-1*offSet)/3600)
	} else {
		offset = fmt.Sprintf("%02d", offSet/3600)
	}
	offset += fmt.Sprintf("%02d", (offSet%3600)/60)

	currentTime := time.Now().In(t.Location())
	//calculating 15 minutes time frame
	diff := currentTime.Minute() - (currentTime.Minute() / 15 * 15)
	begTime := currentTime.Add(time.Duration(-1*diff) * time.Minute)
	endTime := begTime.Add(time.Duration(15) * time.Minute)

	fileName := pmConfig.DestinationDir + "/"
	fileName += filePrefix + begTime.Format(dateFormat) + "." + begTime.Format(timeFormat) + "-" + offset
	fileName += "-" + endTime.Format(timeFormat) + "-" + offset
	fileName += "_" + pmConfig.ForeignID
	counter := 1
	name := fileName
	for fileExists(name) {
		name = fileName + "_" + strconv.Itoa(counter)
		counter++
	}
	fileName = name
	return fileName, nil
}

//checks whether the file exists on the given path
func fileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}
