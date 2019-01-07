/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package formatter

import (
	"bytes"
	"io/ioutil"
	"net"
	"opennmsplugin/config"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var testFMData = `alarmId,dn,alarmNumber,notificationId,originalSeverity,severity,alarmTime,ackStatus,ackTime,ackedBy,eventType,probableCause,alarmText,additionalText1,additionalText2,additionalText3,additionalText4,additionalText5,additionalText6,additionalText7
1234,test_dn,4567,1234,99,1,2017-12-22T14:15:48,1,,,3,0,fault,NA,,NA,,,,`

var testHistoryFMData = `alarmId,dn,alarmTime,eventType,perceivedSeverity,probableCauseCode,specificProblem,alarmText,additionalText1,additionalText2,additionalText3,additionalText4,additionalText5,additionalText6,additionalText7,ackStatus,ackUser,ackTime,alarmState,clearUser,clearTime,rootCauseAlarm,alarmCount,originalSeverity,liftedDn,originalAlarmId
123,test_dn,2018-04-16T02:15:18+03:00,communication,cleared,0,43536,na,,,,,,,,acked,WITHCANCEL,2018-04-16T02:26:01+03:00,cleared,test_user,2018-04-16T02:25:59+03:00,TRUE,1,minor,test_dn,8989790`

func createTestData(filename, content string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	if err != nil {
		return err
	}
	return nil
}

func TestFormatFMData(t *testing.T) {
	fmConfig := config.FMConfig{
		DestinationDir: "./tmp",
		Source:         "source",
		Service:        "NDAC",
		NodeID:         "123",
		Host:           "host",
	}
	fileName := "./fm_data.csv"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	err = os.MkdirAll(fmConfig.DestinationDir, os.ModePerm)
	if err != nil {
		t.Fail()
	}

	tcpAddress := "localhost:3000"
	l, err := net.Listen("tcp", tcpAddress)
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	FormatFMData(fileName, fmConfig, tcpAddress)
	l.Close()
	defer os.RemoveAll(fmConfig.DestinationDir)
	generatedFile := "./tmp/fm_data"
	fileContent, err := ioutil.ReadFile(generatedFile)
	if err != nil || len(fileContent) == 0 {
		t.Fail()
	}
	if !strings.Contains(buf.String(), "FM data "+generatedFile+" pushed to OpenNMS "+tcpAddress) {
		t.Fail()
	}
}

func TestFormatFMDataForHistoryAlarm(t *testing.T) {
	fmConfig := config.FMConfig{
		DestinationDir: "./tmp",
		Source:         "source",
		Service:        "NDAC",
		NodeID:         "123",
		Host:           "host",
	}
	fileName := "./fm_history_data.csv"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	if _, err := os.Stat(fmConfig.DestinationDir); os.IsNotExist(err) {
		os.MkdirAll(fmConfig.DestinationDir, os.ModePerm)
	}

	tcpAddress := "localhost:3000"
	l, err := net.Listen("tcp", tcpAddress)
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	FormatFMData(fileName, fmConfig, tcpAddress)
	l.Close()
	defer os.RemoveAll(fmConfig.DestinationDir)
	generatedFile := "./tmp/fm_history_data"
	fileContent, err := ioutil.ReadFile(generatedFile)
	if err != nil || len(fileContent) == 0 {
		t.Fail()
	}
	if !strings.Contains(buf.String(), "FM data "+generatedFile+" pushed to OpenNMS "+tcpAddress) {
		t.Fail()
	}
}

func TestFormatFMDataWithInvalidFile(t *testing.T) {
	fmConfig := config.FMConfig{
		DestinationDir: "./tmp",
		Source:         "source",
		NodeID:         "123",
		Host:           "host",
	}
	fileName := "./fmdata.csv"
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	FormatFMData(fileName, fmConfig, "")
	if !strings.Contains(buf.String(), "no such file or directory") {
		t.Fail()
	}
}

func TestRetrypushToOpenNMSWithFailedFiles(t *testing.T) {
	nmsAddress := "localhost:3000"
	fileName := "./fm_data.csv"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	failedFmFiles = append(failedFmFiles, fileName)
	l, err := net.Listen("tcp", nmsAddress)
	if err != nil {
		t.Error(err)
	}
	defer l.Close()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	retryTimer = time.NewTimer(10 * time.Millisecond)
	retrypushToOpenNMS(nmsAddress)
	time.Sleep(15 * time.Millisecond)
	if !strings.Contains(buf.String(), "FM data "+fileName+" pushed to OpenNMS "+nmsAddress) {
		t.Fail()
	}
}

func TestFormatFMDataWithoutNMSServer(t *testing.T) {
	fmConfig := config.FMConfig{
		DestinationDir: "./tmp",
		Source:         "source",
		Service:        "NDAC",
		NodeID:         "123",
		Host:           "host",
	}
	fileName := "./fm_data.csv"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)
	if _, err := os.Stat(fmConfig.DestinationDir); os.IsNotExist(err) {
		os.MkdirAll(fmConfig.DestinationDir, os.ModePerm)
	}

	FormatFMData(fileName, fmConfig, "localhost:3000")
	defer os.RemoveAll(fmConfig.DestinationDir)
	if len(failedFmFiles) == 0 || failedFmFiles[0] != "./tmp/fm_data" {
		t.Fail()
	}
}

func TestFormatDateForHistoryAlarm(t *testing.T) {
	formattedTime := formatTime("2018-07-31T16:08:51+03:00", true)
	if formattedTime != "Tuesday, 31 July 2018 13:08:51 o'clock UTC" {
		t.Fail()
	}
}

func TestFormatDateForActiveAlarm(t *testing.T) {
	formattedTime := formatTime("2018-07-31T16:08:51", false)
	if formattedTime != "Tuesday, 31 July 2018 13:08:51 o'clock UTC" {
		t.Fail()
	}
}
