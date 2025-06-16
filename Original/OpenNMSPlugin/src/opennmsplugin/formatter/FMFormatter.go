/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package formatter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"opennmsplugin/config"
	"opennmsplugin/validator"

	log "github.com/sirupsen/logrus"
)

var (
	failedFmFiles []string
	retryTimer    *time.Timer
)

const (
	retryDuration   = 2 * time.Minute
	timeoutDuration = 30 * time.Second

	//time formats
	alarmTimeFormat   = "2006-01-02T15:04:05Z07:00"
	openNMSTimeFormat = "Monday, 02 January 2006 15:04:05 o'clock MST"

	//severity for cleared alarms
	clearedSeverity  = "cleared"
	clearedAlarmType = "ClearedAlarm"
)

type fmdata struct {
	XMLName xml.Name `xml:"log"`
	Events  Events   `xml:"events"`
}

//Events stores the alarms raised
type Events struct {
	Event []Event `xml:"event"`
}

//Event stores a particular alarm data
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

//Params keeps track of additional parameters to be send for a particular alarm
type Params struct {
	Param []Param `xml:"parm"`
}

//Param keeps track of each parameters
type Param struct {
	Name  string     `xml:"parmName"`
	Value ParamValue `xml:"value"`
}

//ParamValue store additional parameter's details
type ParamValue struct {
	Value    string `xml:",chardata"`
	Type     string `xml:"type,attr"`
	Encoding string `xml:"encoding,attr"`
}

//ReceivedFMData stores received FM response from Nokia DAC
type ReceivedFMData struct {
	Source struct {
		EventType             string      `json:"EventType"`
		AlarmState            int         `json:"AlarmState"`
		SpecificProblem       string      `json:"SpecificProblem"`
		LastUpdatedTime       string      `json:"LastUpdatedTime"`
		EdgeID                string      `json:"edge_id"`
		Severity              string      `json:"Severity"`
		NESerialNo            string      `json:"neserialno"`
		NEName                string      `json:"nename"`
		AdditionalText        string      `json:"AdditionalText"`
		EdgeHostname          string      `json:"edge_hostname"`
		ManagedObjectInstance string      `json:"ManagedObjectInstance"`
		ProbableCause         string      `json:"ProbableCause"`
		AlarmText             string      `json:"AlarmText"`
		EventTime             string      `json:"EventTime"`
		NotificationType      string      `json:"NotificationType"`
		AlarmID               interface{} `json:"AlarmIdentifier"`
		EdgeName              string      `json:"edgename"`
		NeHwID                string      `json:"ne_hw_id"`
		Dn                    string      `json:"Dn"`
		NHGID                 string      `json:"nhgid"`
		NHGName               string      `json:"nhgname"`
	} `json:"_source"`
}

//FormatFMData reads csv file containing fm data and converts it to xml and pushes the generated xml file to NMS.
func FormatFMData(filePath string, fmConfig config.FMConfig, openNMSAddress string) {
	log.Infof("Formatting FM file %s", filePath)
	receivedFMData := make([]ReceivedFMData, 0)
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while reading %s", filePath)
		return
	}
	err = json.Unmarshal(contents, &receivedFMData)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}

	fmData := new(fmdata)
	for _, data := range receivedFMData {
		lnbts, lncel := extractLNBTSAndLNCELFromDn(data.Source.Dn)
		params := make([]Param, 0)
		params = append(params, createParam("EventType", data.Source.EventType))
		params = append(params, createParam("AlarmState", strconv.Itoa(data.Source.AlarmState)))
		params = append(params, createParam("SpecificProblem", data.Source.SpecificProblem))
		params = append(params, createParam("EventTime", data.Source.EventTime))
		params = append(params, createParam("edge_id", data.Source.EdgeID))
		params = append(params, createParam("neserialno", data.Source.NESerialNo))
		params = append(params, createParam("AdditionalText", data.Source.AdditionalText))
		params = append(params, createParam("edge_hostname", data.Source.EdgeHostname))
		params = append(params, createParam("ManagedObjectInstance", data.Source.ManagedObjectInstance))
		params = append(params, createParam("ProbableCause", data.Source.ProbableCause))
		params = append(params, createParam("AlarmText", data.Source.AlarmText))
		params = append(params, createParam("nename", data.Source.NEName))
		params = append(params, createParam("edgename", data.Source.EdgeName))
		params = append(params, createParam("AlarmIdentifier", fmt.Sprintf("%v", data.Source.AlarmID)))
		params = append(params, createParam("ne_hw_id", data.Source.NeHwID))
		params = append(params, createParam("Dn", data.Source.Dn))
		params = append(params, createParam("nhgid", data.Source.NHGID))
		params = append(params, createParam("nhgname", data.Source.NHGName))
		params = append(params, createParam(neLnbtsField, lnbts))
		params = append(params, createParam(lncelField, lncel))

		alarmTime := formatTime(data.Source.LastUpdatedTime)
		var severity string
		if data.Source.AlarmState == 0 || data.Source.NotificationType == clearedAlarmType {
			severity = clearedSeverity
		} else {
			severity = data.Source.Severity
		}
		event := Event{UEI: strings.TrimSpace(fmt.Sprintf("%v", data.Source.NotificationType)), Source: fmConfig.Source, NodeID: fmConfig.NodeID, Time: strings.TrimSpace(alarmTime), Host: fmConfig.Host, Severity: strings.TrimSpace(severity), Service: fmConfig.Service}
		event.Params.Param = params
		fmData.Events.Event = append(fmData.Events.Event, event)
	}
	output, _ := xml.MarshalIndent(fmData, "", "  ")

	fileName := filepath.Base(filePath)
	fileName = fileName[0 : len(fileName)-len(filepath.Ext(fileName))]
	fileName = fmConfig.DestinationDir + "/" + fileName
	counter := 1
	name := fileName
	for fileExists(name) {
		name = fileName + "_" + strconv.Itoa(counter)
		counter++
	}
	fileName = name
	file, err := os.Create(fileName)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to create file %s", fileName)
		return
	}
	defer file.Close()
	_, err = file.Write(output)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while writing response to file %s", fileName)
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
	go retryPushToOpenNMS(openNMSAddress)
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
func retryPushToOpenNMS(address string) {
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
	_, err := conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

//Formats the alarm time to OpenNMS time format
func formatTime(alarmTime string) string {
	t, _ := time.Parse(alarmTimeFormat, alarmTime)
	localTimezone := time.Now().Location()
	t = t.In(localTimezone)
	return t.Format(openNMSTimeFormat)
}

func createParam(name string, value string) Param {
	paramValue := ParamValue{Value: strings.TrimSpace(value), Type: "string", Encoding: "text"}
	param := Param{Name: strings.TrimSpace(name), Value: paramValue}
	return param
}
