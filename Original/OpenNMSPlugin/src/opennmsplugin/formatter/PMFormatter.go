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
	"regexp"
	"strconv"
	"time"

	"opennmsplugin/config"
	"opennmsplugin/validator"

	log "github.com/sirupsen/logrus"
)

const (
	dateFormat = "20060102"
	timeFormat = "1504"
	filePrefix = "A"

	//regex for extracting LNBTS_ID and LNCEL_ID from Dn
	dnRegex      = `^.*(?:.*LNBTS-(?P<LNBTS_ID>\d+))(\/)?(?:LNCEL-(?P<LNCEL_ID>\d+))?.*$`
	lncelField   = "LNCEL"
	neLnbtsField = "LNBTS"

	//field name to extract data from PM response file
	sourceField    = "_source"
	eventTimeField = "timestamp"
	dnField        = "Dn"
)

type response struct {
	Data interface{} `json:"data"`
}

//FormatPMData formats PM Data
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

	var pmData interface{}
	err = json.Unmarshal(content, &pmData)
	if err != nil {
		log.Error(err)
		return
	}

	for _, value := range pmData.([]interface{}) {
		source := value.(map[string]interface{})[sourceField].(map[string]interface{})
		metricTime := source[eventTimeField].(string)
		dn := source[dnField].(string)
		lnbts, lncel := extractLNBTSAndLNCELFromDn(dn)
		source[neLnbtsField] = lnbts
		source[lncelField] = lncel

		resp := &response{
			Data: value,
		}
		contents, err := json.Marshal(resp)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Unable to marshal received json response in %s for %s time", filePath, metricTime)
			continue
		}
		fileName, err := renameFile(pmConfig)
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
func renameFile(pmConfig config.PMConfig) (string, error) {
	offset := "0000"
	currentTime := time.Now().UTC()
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

func extractLNBTSAndLNCELFromDn(dn string) (string, string) {
	log.Debugf("Extracting LNBTS and LNCEL from: %s", dn)
	if dn == "" {
		return "", ""
	}
	myExp := regexp.MustCompile(dnRegex)
	match := myExp.FindStringSubmatch(dn)
	if match == nil {
		return "", ""
	}
	result := make(map[string]string)
	for i, name := range myExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result["LNBTS_ID"], result["LNCEL_ID"]
}
