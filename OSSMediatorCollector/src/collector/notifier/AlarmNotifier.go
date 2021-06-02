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
	"os"
	"strings"
	"sync"
	"time"
)

const (
	//Timeout duration for HTTP calls
	timeout             = 60 * time.Second
	alarmConfigFIlePath = "../resources/alarm_notifier.yaml"
)

//AlarmNotifier keeps alarm notification config.
type AlarmNotifier struct {
	MsTeamsWebhook    string         `yaml:"ms_teams_webhook"`
	Filters           []AlarmFilters `yaml:"filters"`
	AlarmSyncDuration int            `yaml:"alarm_sync_duration"`
}

//AlarmFilters stores filters to be applied on alarms before notifying.
type AlarmFilters struct {
	SpecificProblem string   `yaml:"specific_problem"`
	FaultIds        []string `yaml:"fault_ids"`
}

//FMSource struct keeps fm data.
type FMSource struct {
	FmData struct {
		AdditionalText  string `json:"additional_text"`
		AlarmIdentifier string `json:"alarm_identifier"`
		AlarmText       string `json:"alarm_text"`
		EventTime       string `json:"event_time"`
		LastUpdatedTime string `json:"last_updated_time"`
		Severity        string `json:"severity"`
		SpecificProblem string `json:"specific_problem"`
	} `json:"fm_data"`
	FmDataSource struct {
		Dn       string `json:"dn"`
		HwID     string `json:"hw_id"`
		NhgAlias string `json:"nhg_alias"`
	} `json:"fm_data_source"`
}

//RaisedNotification struct to keep track of all the notified alarms.
type RaisedNotification struct {
	alarm            FMSource
	notificationTime time.Time
}

//TeamsMessage forms the body of message to be sent over MS teams.
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

func readAlarmNotifierConfig(txnID uint64) error {
	content, err := ioutil.ReadFile(alarmConfigFIlePath)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error reading YAML file")
		return err
	}

	err = yaml.Unmarshal(content, &alarmNotifier)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error parsing YAML file")
		return err
	}
	return nil
}

//RaiseAlarmNotification alerts about specific alarms configured in resources/alarm_notifier.yaml to MS teams.
func RaiseAlarmNotification(txnID uint64, fmData interface{}) {
	if _, err := os.Stat(alarmConfigFIlePath); os.IsNotExist(err) {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Alarm notifier config not present, skipping alarm notification")
		return
	}
	err := readAlarmNotifierConfig(txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to read alarm notifier config")
		return
	}
	data, _ := json.Marshal(fmData)
	alarmToNotify := getAlarmDetails(txnID, string(data), alarmNotifier.Filters)
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
	var data []FMSource
	var alarmToNotify []FMSource
	err := json.Unmarshal([]byte(fmData), &data)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to parse json fm data")
		return alarmToNotify
	}

	removeOldRaisedAlarms()
	for _, v := range data {
		isValid := checkAlarmFilter(v.FmData.SpecificProblem, v.FmData.AdditionalText, filters)
		if isValid {
			alarmID := getAlarmUniqueID(v)
			_, ok := notifiedAlarms[alarmID]
			if !ok {
				notifiedAlarms[alarmID] = RaisedNotification{
					alarm:            v,
					notificationTime: time.Now(),
				}
				alarmToNotify = append(alarmToNotify, v)
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

	if additionalText == "" {
		return false
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
	msg += fmt.Sprintf("#**Alarm alert for %s network**  \nFollowing alarms have been raised:\n\n---", alarmToNotify[0].FmDataSource.NhgAlias)
	for _, v := range alarmToNotify {
		msg += fmt.Sprintf("\n\n*AlarmID:* **%s**\n\n", v.FmData.AlarmIdentifier)
		msg += fmt.Sprintf("*AlarmText:* **%s**\n\n", v.FmData.AlarmText)
		msg += fmt.Sprintf("*Dn:* **%s**\n\n", v.FmDataSource.Dn)
		msg += fmt.Sprintf("*LastUpdatedTime:* **%s**\n\n", v.FmData.LastUpdatedTime)
		msg += fmt.Sprintf("*Severity:* **%s**\n\n", v.FmData.Severity)
		msg += fmt.Sprintf("*SpecificProblem:* **%s**\n\n", v.FmData.SpecificProblem)
		msg += fmt.Sprintf("*NhgName:* **%s**\n\n", v.FmDataSource.NhgAlias)
		msg += fmt.Sprintf("*AdditionalText:* **%s**\n\n---", v.FmData.AdditionalText)
	}
	return msg
}

func getAlarmUniqueID(alarm FMSource) string {
	id := alarm.FmDataSource.HwID + "_" + alarm.FmDataSource.Dn + "_" + alarm.FmData.AlarmIdentifier + "_" + alarm.FmData.SpecificProblem + "_" + alarm.FmData.EventTime
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
