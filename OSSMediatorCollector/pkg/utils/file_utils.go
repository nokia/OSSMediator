/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package utils

import (
	"collector/pkg/config"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	//CurrentTime
	CurrentTime = time.Now
)

const (
	//File extension for writing response
	fileExtension = ".json"

	fmdataResponseType   = "fmdata"
	pmdataResponseType   = "pmdata"
	nhgResponseType      = "network-hardware-groups"
	gngResponseType      = "generic-network-groups"
	simsResponseType     = "sims"
	orgResponseType      = "organizations"
	accountsResponseType = "accounts"

	//field name to extract data from PM response file
	pmSourceField           = "pm_data_source"
	fmSourceField           = "fm_data"
	eventTimeFieldPM        = "timestamp"
	radioPMEventTimeFieldPM = "end_timestamp"
	eventTimeFieldFM        = "event_time"

	//event time format
	eventTimeFormatBaiCell   = "2006-01-02T15:04:05Z07:00"
	eventTimeFormatNokiaCell = "2006-01-02T15:04:05Z07:00:00"
)

// checks whether file exists, return false if file ids not present.
func fileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

// writes data to file, returns error if file creation or data writing failed.
func writeFile(fileName string, data []byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("file creation failed: %v", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error while writting response to file: %v", err)
	}
	return nil
}

// CreateResponseDirectory creates directory named path, along with any necessary parents.
// If the directory creation fails it will terminate the program.
func CreateResponseDirectory(basePath string, api string) {
	dirPath := basePath + "/" + path.Base(api)
	log.Infof("Creating %s directory", dirPath)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalf("Error while creating %s", dirPath)
	}
}

// WriteResponse writes the data in json format to responseDest directory.
func WriteResponse(user *config.User, api *config.APIConf, data interface{}, id string, txnID uint64, prettyResponse bool) error {
	fileName := path.Base(api.API)
	if fileName == fmdataResponseType || fileName == pmdataResponseType {
		if api.MetricType != "" {
			fileName += "_" + api.MetricType
		}
		if api.Type != "" {
			fileName += "_" + api.Type
		}
		if id != "" {
			fileName += "_" + id
		}
	} else if fileName == nhgResponseType || fileName == gngResponseType {
		fileName += "_" + user.Email
	} else if strings.Contains(fileName, simsResponseType) {
		if id != "" {
			fileName += "_" + id
		} else {
			fileName += "_" + user.Email
		}
	} else if fileName == orgResponseType || fileName == accountsResponseType {
		fileName += "_" + user.Email
		if id != "" {
			fileName += "_" + id
		}
	}

	fileName += "_response_" + strconv.Itoa(int(CurrentTime().Unix()))
	responseDest := user.ResponseDest + "/" + path.Base(api.API)
	fileName = responseDest + "/" + fileName
	counter := 1
	name := fileName
	for fileExists(name + fileExtension) {
		name = fileName + "_" + strconv.Itoa(counter)
		counter++
	}
	fileName = name + fileExtension

	log.WithFields(log.Fields{"tid": txnID}).Infof("Writing response to file %s for %s", fileName, user.Email)
	file, _ := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()
	encoder := json.NewEncoder(file)
	if prettyResponse {
		encoder.SetIndent("", "  ")
	}
	err := encoder.Encode(data)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Writing response to file %s for %s failed", fileName, user.Email)
		return err
	}

	//err = writeFile(fileName, formattedData)
	//if err != nil {
	//	log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Writing response to file %s for %s failed", fileName, user.Email)
	//	return err
	//}
	return nil
}

// retrieves the last received metric time from file per API
func getLastReceivedDataTime(user *config.User, api *config.APIConf, nhgID string) string {
	//Reading start time value from file
	fileName := "checkpoints/" + path.Base(api.API)
	if api.MetricType != "" {
		fileName = fileName + "_" + api.MetricType
	}
	if api.Type != "" {
		fileName = fileName + "_" + api.Type
	}
	fileName = fileName + "_" + user.Email
	if nhgID != "" {
		fileName = fileName + "_" + nhgID
	}
	data, err := os.ReadFile(fileName)
	if err != nil || len(data) == 0 {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// StoreLastReceivedDataTime writes the last received metric's event time to a file so that next time that time stamp will be used as start_time for api calls.
// Creates file for each user and each API.
// returns error if writing to or reading from the response directory fails.
func StoreLastReceivedDataTime(user *config.User, data interface{}, api *config.APIConf, nhgID string, txnID uint64) error {
	var fieldName, source string
	baseAPIPath := path.Base(api.API)
	if baseAPIPath == fmdataResponseType {
		source = fmSourceField
		fieldName = eventTimeFieldFM
	} else if baseAPIPath == pmdataResponseType {
		if api.MetricType == "RADIO" {
			fieldName = radioPMEventTimeFieldPM
		} else {
			fieldName = eventTimeFieldPM
		}
		source = pmSourceField
	}

	receivedData := make(map[string]struct{})
	for _, value := range data.([]interface{}) {
		eventTime := value.(map[string]interface{})[source].(map[string]interface{})[fieldName]
		if eventTime != nil {
			metricTime := eventTime.(string)
			receivedData[metricTime] = struct{}{}
		}
	}
	data = nil
	var eventTimes []time.Time
	for key := range receivedData {
		eventTime, err := time.Parse(eventTimeFormatNokiaCell, key)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "event_time": eventTime}).Errorf("Unable to parse event time using Nokia cell or BaiCell time format")
			continue
		}
		eventTimes = append(eventTimes, eventTime)
	}
	lastReceivedTime := getLastReceivedDataTime(user, api, nhgID)
	if lastReceivedTime != "" {
		t, err := time.Parse(time.RFC3339, lastReceivedTime)
		if err == nil {
			eventTimes = append(eventTimes, t)
		}
	}
	sort.Slice(eventTimes, func(i, j int) bool { return eventTimes[i].Before(eventTimes[j]) })
	latestTime := truncateSeconds(eventTimes[len(eventTimes)-1]).Format(time.RFC3339)

	fileName := "checkpoints/" + path.Base(api.API)
	if api.MetricType != "" {
		fileName = fileName + "_" + api.MetricType
	}
	if api.Type != "" {
		fileName = fileName + "_" + api.Type
	}
	fileName = fileName + "_" + user.Email
	if nhgID != "" {
		fileName = fileName + "_" + nhgID
	}
	log.WithFields(log.Fields{"tid": txnID}).Debug("Writing to ", fileName)
	err := writeFile(fileName, []byte(latestTime))
	if err != nil {
		return fmt.Errorf("unable to write last received data time, error: %v", err)
	}
	return nil
}
