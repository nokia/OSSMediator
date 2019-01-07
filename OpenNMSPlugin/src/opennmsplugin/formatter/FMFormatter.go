/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package formatter

import (
	"bufio"
	"encoding/csv"
	"encoding/xml"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"opennmsplugin/config"
	"opennmsplugin/validator"

	log "github.com/sirupsen/logrus"
)

var (
	failedFmFiles []string
	retryTimer    *time.Timer
	fmSeverity    = map[string]string{
		"1": "critical",
		"2": "major",
		"3": "minor",
		"4": "warning",
	}
)

const (
	retryDuration   = 2 * time.Minute
	timeoutDuration = 30 * time.Second

	//time formats
	activeAlarmTimeFormat  = "2006-01-02T15:04:05"
	historyAlarmTimeFormat = "2006-01-02T15:04:05-07:00"
	openNMSTimeFormat      = "Monday, 02 January 2006 15:04:05 o'clock MST"
	netACTLocation         = "Europe/Moscow"

	//field name in FM data
	severity          = "severity"
	perceivedSeverity = "perceivedSeverity"
	alarmNumber       = "alarmNumber"
	specificProblem   = "specificProblem"
	alarmTime         = "alarmTime"
)

type fmdata struct {
	XMLName xml.Name `xml:"log"`
	Events  Events   `xml:"events"`
}

type Events struct {
	Event []Event `xml:"event"`
}

type Event struct {
	UEI      string `xml:"uei"`
	Source   string `xml:"source"`
	Service  string `xml:"service"`
	NodeID   string `xml:"nodeid"`
	Time     string `xml:"time"`
	Host     string `xml:"host"`
	Severity string `xml:"severity"`
	Params   Params `xml:"parms"`
}

type Params struct {
	Param []Param `xml:"parm"`
}

type Param struct {
	Name  string     `xml:"parmName"`
	Value ParamValue `xml:"value"`
}

type ParamValue struct {
	Value    string `xml:",chardata"`
	Type     string `xml:"type,attr"`
	Encoding string `xml:"encoding,attr"`
}

//FormatFMData reads csv file containing fm data and converts it to xml and pushes the generated xml file to NMS.
func FormatFMData(filePath string, fmConfig config.FMConfig, openNMSAddress string) {
	log.Infof("Formatting FM file %s", filePath)
	//read csv
	csvFile, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading %s", filePath)
		return
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	defer csvFile.Close()
	//removing source file
	defer os.Remove(filePath)

	fmData := new(fmdata)
	lines, err := reader.ReadAll()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading %s", filePath)
		return
	}
	var headers []string
	var alarmSeverityIndx, alarmNoIndx, alarmTimeIndx int

	if len(lines) <= 1 {
		log.Warningf("%s doen't have FM data", filePath)
		return
	}

	isHistoryAlarm := false
	for i, line := range lines {
		if i == 0 {
			headers = line
			alarmSeverityIndx = findIndex(line, severity)
			if alarmSeverityIndx == -1 {
				alarmSeverityIndx = findIndex(line, perceivedSeverity)
				isHistoryAlarm = true
			}

			alarmNoIndx = findIndex(line, alarmNumber)
			if alarmNoIndx == -1 {
				alarmNoIndx = findIndex(line, specificProblem)
			}
			alarmTimeIndx = findIndex(line, alarmTime)
			if alarmNoIndx == -1 || alarmSeverityIndx == -1 || alarmTimeIndx == -1 {
				return
			}
			continue
		}

		params := make([]Param, 0)
		for index, value := range line {
			if index == alarmNoIndx || index == alarmTimeIndx || index == alarmSeverityIndx {
				continue
			}
			paramValue := ParamValue{Value: strings.TrimSpace(value), Type: "string", Encoding: "text"}
			param := Param{Name: strings.TrimSpace(headers[index]), Value: paramValue}
			params = append(params, param)
		}
		var severity string
		if isHistoryAlarm {
			severity = line[alarmSeverityIndx]
		} else {
			severity = fmSeverity[line[alarmSeverityIndx]]
		}

		alarmTime := formatTime(line[alarmTimeIndx], isHistoryAlarm)
		event := Event{UEI: strings.TrimSpace(line[alarmNoIndx]), Source: fmConfig.Source, NodeID: fmConfig.NodeID, Time: strings.TrimSpace(alarmTime), Host: fmConfig.Host, Severity: strings.TrimSpace(severity), Service: fmConfig.Service}
		event.Params.Param = params
		fmData.Events.Event = append(fmData.Events.Event, event)
	}
	output, _ := xml.MarshalIndent(fmData, "", "  ")

	fileName := filepath.Base(csvFile.Name())
	fileName = fileName[0 : len(fileName)-len(filepath.Ext(fileName))]
	fileName = fmConfig.DestinationDir + "/" + fileName
	file, err := os.Create(fileName)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to create file %s", fileName)
		return
	}
	defer file.Close()
	_, err = file.Write(output)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while writting response to file %s", fileName)
		return
	}

	//Validate the xml
	err = validator.ValidateXML(fileName)
	if err != nil {
		log.Errorf("Invalid xml generated from %s", filePath)
		os.Remove(fileName)
		return
	}
	log.Infof("Formatted FM data written to %s", fileName)
	pushToOpenNMS(openNMSAddress, fileName)
	//retry push to OpeNMS if push failed
	go retrypushToOpenNMS(openNMSAddress)
}

//Establishes the TCP connection with OpenNMS address.
//returns error if connection fails.
func connectOpenNMS(addr string) (net.Conn, error) {
	var connection net.Conn
	var attempt int
	var err error
	for {
		dialer := net.Dialer{Timeout: timeoutDuration}
		connection, err = dialer.Dial("tcp", addr)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Info("OpenNMS Dial faied, retrying...")
			attempt++
			if attempt >= 3 {
				break
			}
			continue
		}
		break
	}
	return connection, err
}

func findIndex(line []string, key string) int {
	for i, n := range line {
		if key == n {
			return i
		}
	}
	return -1
}

//Pushes fm files to OpenNMS via TCP connection, if it is unable to connect to
//the address the it adds the files to a list and starts a timer to retry the push operation.
func pushToOpenNMS(address string, filename string) {
	log.Infof("Pushing %s FM data to OpenNMS %s", filename, address)
	conn, err := connectOpenNMS(address)
	if err != nil {
		log.Error("Unable to connect to OpenNMS after three attempts, fm data will be pushed later...")
		failedFmFiles = append(failedFmFiles, filename)
	} else {
		err = pushFile(conn, filename)
		if err != nil {
			failedFmFiles = append(failedFmFiles, filename)
			log.WithFields(log.Fields{"error": err}).Errorf("Write to OpenNMS failed for file %s, it will be pushed again", filename)
		} else {
			log.Infof("FM data %s pushed to OpenNMS %s", filename, address)
			indx := findIndex(failedFmFiles, filename)
			if indx != -1 {
				failedFmFiles = append(failedFmFiles[:indx], failedFmFiles[indx+1:]...)
			}
		}
		conn.Close()
	}
}

//Pushes the failed fm files to OpenNMS after expiry of retryTimer.
func retrypushToOpenNMS(address string) {
	if len(failedFmFiles) == 0 {
		return
	}
	if retryTimer == nil {
		retryTimer = time.NewTimer(retryDuration)
	}
	for {
		<-retryTimer.C
		log.Infof("Retrying to push FM data to OpenNMS %s", address)
		_, err := connectOpenNMS(address)
		if err != nil {
			retryTimer.Reset(retryDuration)
		} else {
			for _, file := range failedFmFiles {
				pushToOpenNMS(address, file)
			}
			retryTimer = nil
			return
		}
	}
}

//writes the fm file to OpenNMS
func pushFile(conn net.Conn, filename string) error {
	b, _ := ioutil.ReadFile(filename)
	_, err := conn.Write([]byte(b))
	if err != nil {
		return err
	}
	return nil
}

//Formats the alarm time to OpenNMS time format
func formatTime(alarmTime string, isHistoryAlarm bool) string {
	ndacFormat := activeAlarmTimeFormat
	if isHistoryAlarm {
		ndacFormat = historyAlarmTimeFormat
	}

	t, _ := time.Parse(ndacFormat, alarmTime)
	netACTLoc, _ := time.LoadLocation(netACTLocation)
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, netACTLoc)
	localTimezone := time.Now().Location()
	t = t.In(localTimezone)
	formattedTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, localTimezone)
	return formattedTime.Format(openNMSTimeFormat)
}
