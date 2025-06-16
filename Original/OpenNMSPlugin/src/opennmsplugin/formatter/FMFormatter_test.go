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

var testFMData = `[
    {
       "_index":"test_index",
       "_type":"",
       "_id":"12345",
       "_source": {
          "Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
          "EventType":"Processing Error Alarm",
          "AlarmState":0,
          "SpecificProblem":"Time synchronization failed",
          "LastUpdatedTime":"2018-09-23T16:04:06Z",
          "edge_id":"edge_id",
          "Severity":"Minor",
          "NESerialNo":"98675",
          "NeName":"",
          "AdditionalText":"Unspecified",
          "edge_hostname":"localhost",
          "ProbableCause":"Time synchronization failed",
          "AlarmText":"Time synchronization failed",
          "EventTime":"2018-09-23T21:33:35+05:30",
          "NotificationType":"NewAlarm",
          "AlarmIdentifier":"11191",
          "EdgeName":"",
          "nhgid":"123456",
		  "nhgname":"test_nhg"
       }
    },
    {
       "_index":"test_index",
       "_type":"",
       "_id":"12345",
       "_source": {
	      "Dn": "NE-MRBTS-1/NE-LNBTS-1/MCC-244/MNC-09",
          "EventType":"Processing Error Alarm",
          "AlarmState":0,
          "SpecificProblem":"Time synchronization failed",
          "LastUpdatedTime":"2018-09-24T10:58:06Z",
          "edge_id":"edge_id",
          "Severity":"Minor",
          "NESerialNo":"564357",
          "NeName":"",
          "AdditionalText":"Unspecified",
          "edge_hostname":"localhost",
          "AlarmText":"Time synchronization failed",
          "EventTime":"2018-09-24T16:28:01+05:30",
          "NotificationType":"ClearedAlarm",
          "AlarmIdentifier":"11191",
          "EdgeName":"",
          "nhgid":"65477",
		  "nhgname":"test_nhg"
       }
    }
 ]`

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
	fileName := "./fm_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	err = os.MkdirAll(fmConfig.DestinationDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	tcpAddress := "localhost:8888"
	l, err := net.Listen("tcp", tcpAddress)
	if err != nil {
		t.Error(err)
	} else {
		defer l.Close()
	}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	FormatFMData(fileName, fmConfig, tcpAddress)

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

func TestFormatFMDataWithInvalidFile(t *testing.T) {
	fmConfig := config.FMConfig{
		DestinationDir: "./tmp",
		Source:         "source",
		NodeID:         "123",
		Host:           "host",
	}
	fileName := "./tmp.json"
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

func TestRetrypushToOpenNMSWithFailedFilesPush(t *testing.T) {
	nmsAddress := "localhost:8888"
	fileName := "./fm_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	failedFmFiles = append(failedFmFiles, fileName)
	l, err := net.Listen("tcp", nmsAddress)
	if err != nil {
		t.Error(err)
	} else {
		defer l.Close()
	}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	retryTimer = time.NewTimer(10 * time.Millisecond)
	retryPushToOpenNMS(nmsAddress)
	time.Sleep(15 * time.Millisecond)
	t.Log(buf.String())
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
	fileName := "./fm_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	err = os.MkdirAll(fmConfig.DestinationDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	FormatFMData(fileName, fmConfig, "localhost:8888")
	defer os.RemoveAll(fmConfig.DestinationDir)
	if len(failedFmFiles) == 0 || failedFmFiles[0] != "./tmp/fm_data" {
		t.Fail()
	}
}

func TestFormatDate(t *testing.T) {
	formattedTime := formatTime("2018-09-24T05:01:52Z")
	t.Log(formattedTime)
	if formattedTime != "Monday, 24 September 2018 05:01:52 o'clock UTC" {
		t.Fail()
	}
}
