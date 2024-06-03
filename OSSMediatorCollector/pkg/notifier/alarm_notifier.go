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
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	//Timeout duration for HTTP calls
	timeout             = 60 * time.Second
	alarmConfigFIlePath = "../resources/alarm_notifier.yaml"
	//message format
	msTeamsMsgFormat = "ms_teams"
	jsonMsgFormat    = "json"
)

// AlarmNotifier keeps alarm notification config.
type AlarmNotifier struct {
	WebhookURL        string              `yaml:"webhook_url"`
	RadioAlarmFilters []RadioAlarmFilters `yaml:"radio_alarm_filters"`
	DACAlarmFilters   []AlarmIDFilters    `yaml:"dac_alarm_filters"`
	COREAlarmFilters  []AlarmIDFilters    `yaml:"core_alarm_filters"`
	AlarmSyncDuration int                 `yaml:"alarm_sync_duration"`
	GroupEvents       bool                `yaml:"group_events"`
	NotifyClearEvents bool                `yaml:"notify_clear_event"`
	MessageFormat     string              `yaml:"message_format"`
}

// AlarmIDFilters stores alarm_id to be applied on dac/core alarms before notifying.
type AlarmIDFilters struct {
	AlarmID string `yaml:"alarm_id"`
}

// RadioAlarmFilters stores filters to be applied on radio alarms before notifying.
type RadioAlarmFilters struct {
	SpecificProblem string   `yaml:"specific_problem"`
	FaultIds        []string `yaml:"fault_ids"`
}

// FMSource struct keeps fm data.
type FMSource struct {
	FmData struct {
		AdditionalText   string `json:"additional_text"`
		AlarmIdentifier  string `json:"alarm_identifier"`
		AlarmState       string `json:"alarm_state"`
		AlarmText        string `json:"alarm_text"`
		ClearAlarmTime   string `json:"clear_alarm_time"`
		ClearText        string `json:"clear_text"`
		EventTime        string `json:"event_time"`
		EventType        string `json:"event_type"`
		FaultID          string `json:"fault_id"`
		LastUpdatedTime  string `json:"last_updated_time"`
		NotificationType string `json:"notification_type"`
		ProbableCause    string `json:"probable_cause"`
		Severity         string `json:"severity"`
		SpecificProblem  string `json:"specific_problem"`
	} `json:"fm_data"`
	FmDataSource struct {
		ApID         string `json:"ap_id"`
		AppID        string `json:"app_id"`
		AppName      string `json:"app_name"`
		ClusterID    string `json:"cluster_id"`
		DcID         string `json:"dc_id"`
		DcName       string `json:"dc_name"`
		Dn           string `json:"dn"`
		EdgeAlias    string `json:"edge_alias"`
		EdgeHostname string `json:"edge_hostname"`
		EdgeID       string `json:"edge_id"`
		HwAlias      string `json:"hw_alias"`
		HwID         string `json:"hw_id"`
		NhgAlias     string `json:"nhg_alias"`
		NhgID        string `json:"nhg_id"`
		SerialNo     string `json:"serial_no"`
		SliceID      string `json:"slice_id"`
		Technology   string `json:"technology"`
	} `json:"fm_data_source"`
}

// RaisedNotification struct to keep track of all the notified alarms.
type RaisedNotification struct {
	alarm            FMSource
	notificationTime time.Time
}

// TeamsMessage forms the body of message to be sent over MS teams.
type TeamsMessage struct {
	Text       string `json:"text"`
	TextFormat string `json:"textFormat,omitempty"`
}

var (
	once           sync.Once
	netClient      *http.Client
	alarmNotifier  AlarmNotifier
	notifiedAlarms = make(map[string]RaisedNotification)
	mux            sync.Mutex
)

func readAlarmNotifierConfig(txnID uint64) error {
	content, err := os.ReadFile(alarmConfigFIlePath)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error reading YAML file")
		return err
	}

	err = yaml.Unmarshal(content, &alarmNotifier)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Error parsing YAML file")
		return err
	}
	re := regexp.MustCompile(`^(ms_teams|json)$`)
	if !re.MatchString(alarmNotifier.MessageFormat) {
		log.WithFields(log.Fields{"tid": txnID}).Errorf("Invalid message format, message_format should be ms_team/json")
		return errors.New("invalid message format, message_format should be ms_team/json")
	}
	return nil
}

// RaiseAlarmNotification alerts about specific alarms configured in resources/alarm_notifier.yaml to MS teams.
func RaiseAlarmNotification(txnID uint64, fmData interface{}, metricType string, eventType string) {
	if _, err := os.Stat(alarmConfigFIlePath); os.IsNotExist(err) {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Alarm notifier config not present, skipping alarm notification")
		return
	}
	err := readAlarmNotifierConfig(txnID)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to read alarm notifier config")
		return
	}

	if eventType == "HISTORY" && !alarmNotifier.NotifyClearEvents {
		log.WithFields(log.Fields{"tid": txnID}).Infof("History alarm notifier not enabled, skipping alarm notification for History alarms")
		return
	}

	data, _ := json.Marshal(fmData)
	alarmToNotify := getAlarmDetails(txnID, string(data), metricType)
	if len(alarmToNotify) == 0 {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Found no alarms to notify")
		return
	}

	var message []byte
	if alarmNotifier.GroupEvents {
		if alarmNotifier.MessageFormat == msTeamsMsgFormat {
			message = formMSTeamsMessage(txnID, alarmToNotify)
		} else if alarmNotifier.MessageFormat == jsonMsgFormat {
			message = formJSONMessage(txnID, alarmToNotify)
		}
		if message != nil {
			pushToWebHook(txnID, message)
		}
	} else {
		for _, alarm := range alarmToNotify {
			if alarmNotifier.MessageFormat == msTeamsMsgFormat {
				message = formMSTeamsMessage(txnID, []FMSource{alarm})
			} else if alarmNotifier.MessageFormat == jsonMsgFormat {
				message = formJSONMessage(txnID, []FMSource{alarm})
			}
			if message != nil {
				pushToWebHook(txnID, message)
			}
		}
	}
}

func pushToWebHook(txnID uint64, message []byte) {
	request, err := http.NewRequest("POST", alarmNotifier.WebhookURL, bytes.NewBuffer(message))
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
	log.WithFields(log.Fields{"tid": txnID}).Infof("Alarms notified")
}

func getAlarmDetails(txnID uint64, fmData string, metricType string) []FMSource {
	var data []FMSource
	var alarmToNotify []FMSource
	err := json.Unmarshal([]byte(fmData), &data)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID, "error": err}).Errorf("Unable to parse json fm data")
		return alarmToNotify
	}

	removeOldRaisedAlarms()
	for _, v := range data {
		var isValid bool
		if metricType == "RADIO" {
			isValid = checkRadioAlarmFilter(v.FmData.SpecificProblem, v.FmData.AdditionalText, alarmNotifier.RadioAlarmFilters)
		} else if metricType == "DAC" {
			isValid = checkAlarmIDFilter(v.FmData.AlarmIdentifier, alarmNotifier.DACAlarmFilters)
		} else if metricType == "CORE" {
			isValid = checkAlarmIDFilter(v.FmData.AlarmIdentifier, alarmNotifier.COREAlarmFilters)
		}

		if isValid {
			alarmID := getAlarmUniqueID(v, metricType)
			mux.Lock()
			_, ok := notifiedAlarms[alarmID]
			if !ok {
				notifiedAlarms[alarmID] = RaisedNotification{
					alarm:            v,
					notificationTime: time.Now(),
				}
				alarmToNotify = append(alarmToNotify, v)
			}
			mux.Unlock()
		}
	}
	return alarmToNotify
}

func checkAlarmIDFilter(alarmID string, alarmIDFilters []AlarmIDFilters) bool {
	for _, v := range alarmIDFilters {
		if v.AlarmID == alarmID || v.AlarmID == "*" {
			return true
		}
	}
	return false
}

func checkRadioAlarmFilter(specificProblem string, additionalText string, filters []RadioAlarmFilters) bool {
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
		if v == specificProblem || v == "*" {
			return true
		}
	}
	return false
}

func formJSONMessage(txnID uint64, alarmToNotify []FMSource) []byte {
	alarmData, err := json.Marshal(alarmToNotify)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Unable to marshal message")
		return nil
	}
	message := TeamsMessage{Text: string(alarmData)}
	data, err := json.Marshal(message)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Unable to marshal message")
		return nil
	}
	return data
}

func formMSTeamsMessage(txnID uint64, alarmToNotify []FMSource) []byte {
	var msg string
	if alarmToNotify[0].FmDataSource.NhgAlias != "" {
		msg += fmt.Sprintf("#**Alarm alert for %s network**  \nFollowing alarms have been raised:\n\n---", strings.TrimSpace(alarmToNotify[0].FmDataSource.NhgAlias))
	} else {
		msg += fmt.Sprintf("#**Alarm alert for %s network**  \nFollowing alarms have been raised:\n\n---", strings.TrimSpace(alarmToNotify[0].FmDataSource.NhgID))
	}

	for _, v := range alarmToNotify {
		msg += fmt.Sprintf("\n\n*AlarmID:* **%s**\n\n", strings.TrimSpace(v.FmData.AlarmIdentifier))
		msg += fmt.Sprintf("*AlarmText:* **%s**\n\n", strings.TrimSpace(v.FmData.AlarmText))
		if v.FmDataSource.Dn != "" {
			msg += fmt.Sprintf("*Dn:* **%s**\n\n", strings.TrimSpace(v.FmDataSource.Dn))
		}
		msg += fmt.Sprintf("*AlarmState:* **%s**\n\n", strings.TrimSpace(v.FmData.AlarmState))
		msg += fmt.Sprintf("*LastUpdatedTime:* **%s**\n\n", strings.TrimSpace(v.FmData.LastUpdatedTime))
		msg += fmt.Sprintf("*Severity:* **%s**\n\n", strings.TrimSpace(v.FmData.Severity))
		if v.FmData.SpecificProblem != "" {
			msg += fmt.Sprintf("*SpecificProblem:* **%s**\n\n", strings.TrimSpace(v.FmData.SpecificProblem))
		}
		if v.FmDataSource.NhgAlias != "" {
			msg += fmt.Sprintf("*NhgName:* **%s**\n\n", strings.TrimSpace(v.FmDataSource.NhgAlias))
		} else {
			msg += fmt.Sprintf("*NhgID:* **%s**\n\n", strings.TrimSpace(v.FmDataSource.NhgID))
		}

		if v.FmData.AdditionalText != "" {
			msg += fmt.Sprintf("*AdditionalText:* **%s**\n\n", strings.TrimSpace(v.FmData.AdditionalText))
		}
		msg += fmt.Sprintf("---")
	}
	message := TeamsMessage{Text: msg, TextFormat: "markdown"}
	alarmData, err := json.Marshal(message)
	if err != nil {
		log.WithFields(log.Fields{"tid": txnID}).Debugf("Unable to marshal message")
		return nil
	}
	return alarmData
}

func getAlarmUniqueID(alarm FMSource, metricType string) string {
	var keys []string
	if metricType == "RADIO" {
		keys = []string{alarm.FmDataSource.HwID, alarm.FmDataSource.Dn, alarm.FmData.AlarmIdentifier, alarm.FmData.SpecificProblem, alarm.FmData.EventTime}
	} else if metricType == "DAC" {
		keys = []string{alarm.FmDataSource.Dn, alarm.FmData.AlarmIdentifier, alarm.FmData.EventTime}
	} else if metricType == "CORE" {
		keys = []string{alarm.FmDataSource.NhgID, alarm.FmDataSource.EdgeID, alarm.FmData.AlarmIdentifier, alarm.FmData.EventTime}
	}
	id := strings.Join(keys, "_")
	return id
}

func removeOldRaisedAlarms() {
	cleanupDuration := time.Duration(alarmNotifier.AlarmSyncDuration) * time.Minute
	mux.Lock()
	for k, v := range notifiedAlarms {
		if time.Now().Sub(v.notificationTime) > cleanupDuration {
			delete(notifiedAlarms, k)
		}
	}
	mux.Unlock()
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
