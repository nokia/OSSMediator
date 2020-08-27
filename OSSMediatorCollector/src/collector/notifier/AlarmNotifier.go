/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	//Timeout duration for HTTP calls
	timeout     = 60 * time.Second
	alarmConfig = "../resources/alarm_notifier.yaml"
)

type AlarmNotifier struct {
	MsTeamsWebhook    string         `yaml:"ms_teams_webhook"`
	Filters           []AlarmFilters `yaml:"filters"`
	AlarmSyncDuration int            `yaml:"alarm_sync_duration"`
}

type AlarmFilters struct {
	SpecificProblem string   `yaml:"specific_problem"`
	FaultIds        []string `yaml:"fault_ids"`
}

type FMResponse []struct {
	Source FMSource `json:"_source"`
}

type FMSource struct {
	AlarmIdentifier string `json:"AlarmIdentifier"`
	AdditionalText  string `json:"AdditionalText"`
	AlarmText       string `json:"AlarmText"`
	Dn              string `json:"Dn"`
	EventTime       string `json:"EventTime"`
	LastUpdatedTime string `json:"LastUpdatedTime"`
	Severity        string `json:"Severity"`
	SpecificProblem string `json:"SpecificProblem"`
	NhgName         string `json:"nhgname"`
	NeHwID          string `json:"ne_hw_id"`
}

type RaisedNotification struct {
	alarm            FMSource
	notificationTime time.Time
}

type TeamsMessage struct {
	Text       string `json:"text"`
	TextFormat string `json:"textFormat"`
}

var (
	once           sync.Once
	netClient      *http.Client
	alarmNotifier  AlarmNotifier
	notifiedAlarms = make(map[string]RaisedNotification)
)

func readAlarmNotifierConfig(txnID uint64) {
	content, err := ioutil.ReadFile(alarmConfig)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error reading YAML file")
		return
	}

	err = yaml.Unmarshal(content, &alarmNotifier)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error parsing YAML file")
	}
}

func RaiseAlarmNotification(txnID uint64, fmData string) {
	readAlarmNotifierConfig(txnID)
	alarmToNotify := getAlarmDetails(txnID, fmData, alarmNotifier.Filters)
	if len(alarmToNotify) == 0 {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Found no alarms to notify")
		return
	}
	message := formMessage(alarmToNotify)
	msg := TeamsMessage{Text: message, TextFormat: "markdown"}
	b, err := json.Marshal(msg)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Unable to marshal message")
		return
	}
	request, err := http.NewRequest("POST", alarmNotifier.MsTeamsWebhook, bytes.NewBuffer(b))
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to send alarm notification")
		return
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := newNetClient().Do(request)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to send alarm notification")
		return
	}
	if response.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"tid": txnID}).Errorf("Alarm notification received response' status code: %d, status: %s", response.StatusCode, response.Status)
		return
	}

	response.Body.Close()
}

func getAlarmDetails(txnID uint64, fmData string, filters []AlarmFilters) []FMSource {
	var data FMResponse
	var alarmToNotify []FMSource
	err := json.Unmarshal([]byte(fmData), &data)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to parse json fm data")
		return alarmToNotify
	}

	removeOldRaisedAlarms()
	for _, v := range data {
		isValid := checkAlarmFilter(v.Source.SpecificProblem, v.Source.AdditionalText, filters)
		if isValid {
			alarmID := getAlarmUniqueID(v.Source)
			_, ok := notifiedAlarms[alarmID]
			if !ok {
				notifiedAlarms[alarmID] = RaisedNotification{
					alarm:            v.Source,
					notificationTime: time.Now(),
				}
				alarmToNotify = append(alarmToNotify, v.Source)
			}
		}
	}
	return alarmToNotify
}

func checkAlarmFilter(specificProblem string, additionalText string, filters []AlarmFilters) bool {
	var spProbs []string
	for _, v := range filters {
		spProbs = append(spProbs, v.SpecificProblem)
	}

	if !checkSpecificProblem(specificProblem, spProbs) {
		return false
	}

	var faultIDs []string
	for _, v := range filters {
		if v.SpecificProblem == specificProblem {
			faultIDs = v.FaultIds
			break
		}
	}

	if len(faultIDs) == 0 {
		return true
	}

	alarmFault := strings.Split(additionalText, ";")[1]
	if alarmFault == "" {
		return false
	}
	for _, faultID := range faultIDs {
		if faultID == alarmFault {
			return true
		}
	}
	return false
}

func checkSpecificProblem(specificProblem string, alarmSpecificProblems []string) bool {
	for _, v := range alarmSpecificProblems {
		if v == specificProblem {
			return true
		}
	}
	return false
}

func formMessage(alarmToNotify []FMSource) string {
	var msg string
	msg += fmt.Sprintf("#**Alarm alert for %s network**  \nFollowing alarms have been raised:\n\n---", alarmToNotify[0].NhgName)
	for _, v := range alarmToNotify {
		msg += fmt.Sprintf("\n\n*AlarmID:* **%s**\n\n", v.AlarmIdentifier)
		msg += fmt.Sprintf("*AlarmText:* **%s**\n\n", v.AlarmText)
		msg += fmt.Sprintf("*Dn:* **%s**\n\n", v.Dn)
		msg += fmt.Sprintf("*LastUpdatedTime:* **%s**\n\n", v.LastUpdatedTime)
		msg += fmt.Sprintf("*Severity:* **%s**\n\n", v.Severity)
		msg += fmt.Sprintf("*SpecificProblem:* **%s**\n\n", v.SpecificProblem)
		msg += fmt.Sprintf("*NhgName:* **%s**\n\n", v.NhgName)
		msg += fmt.Sprintf("*AdditionalText:* **%s**\n\n---", v.AdditionalText)
	}
	return msg
}

func getAlarmUniqueID(alarm FMSource) string {
	id := alarm.NeHwID + "_" + alarm.Dn + "_" + alarm.AlarmIdentifier + "_" + alarm.SpecificProblem + "_" + alarm.EventTime
	return id
}

func removeOldRaisedAlarms() {
	cleanupDuration := time.Duration(alarmNotifier.AlarmSyncDuration) * time.Minute
	for k, v := range notifiedAlarms {
		if time.Now().Sub(v.notificationTime) > cleanupDuration {
			delete(notifiedAlarms, k)
		}
	}
}

func newNetClient() *http.Client {
	once.Do(func() {
		netTransport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		netClient = &http.Client{
			Transport: netTransport,
			Timeout:   timeout,
		}
	})

	return netClient
}
