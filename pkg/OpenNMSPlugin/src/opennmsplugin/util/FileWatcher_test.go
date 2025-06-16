/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"opennmsplugin/config"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

func TestAddWatcher(t *testing.T) {
	tmpdir1 := "./tmp1"
	tmpdir2 := "./tmp2"
	users := []config.UserConf{
		{
			SourceDir: tmpdir1,
		},
		{
			SourceDir: tmpdir2,
		},
	}
	conf := config.Config{
		UsersConf: users,
	}
	if _, err := os.Stat(tmpdir1); os.IsNotExist(err) {
		os.MkdirAll(tmpdir1+"/pm", os.ModePerm)
		os.MkdirAll(tmpdir1+"/fm", os.ModePerm)
	}
	if _, err := os.Stat(tmpdir2); os.IsNotExist(err) {
		os.MkdirAll(tmpdir2+"/pm", os.ModePerm)
		os.MkdirAll(tmpdir2+"/fm", os.ModePerm)
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	err := AddWatcher(conf)
	defer os.RemoveAll(tmpdir1)
	defer os.RemoveAll(tmpdir2)
	if err != nil || !strings.Contains(buf.String(), "Added watcher to ./tmp1/fm") || !strings.Contains(buf.String(), "Added watcher to ./tmp1/pm") || !strings.Contains(buf.String(), "Added watcher to ./tmp2/pm") || !strings.Contains(buf.String(), "Added watcher to ./tmp2/fm") {
		t.Fail()
	}
}

func TestAddWatcherWithEmptySourceDir(t *testing.T) {
	users := []config.UserConf{
		{
			SourceDir: "",
		},
	}
	conf := config.Config{
		UsersConf: users,
	}
	err := AddWatcher(conf)
	if err == nil && strings.Contains(err.Error(), "Source directory path can't be empty") {
		t.Fail()
	}
}

func TestAddWatcherWithClosedWatcherInstance(t *testing.T) {
	tmpdir := "./tmp"
	users := []config.UserConf{
		{
			SourceDir: tmpdir,
		},
	}
	conf := config.Config{
		UsersConf: users,
	}
	if _, err := os.Stat(tmpdir); os.IsNotExist(err) {
		os.MkdirAll(tmpdir+"/pm", os.ModePerm)
		os.MkdirAll(tmpdir+"/fm", os.ModePerm)
	}
	defer os.RemoveAll(tmpdir)

	watcher, _ = fsnotify.NewWatcher()
	watcher.Close()
	err := AddWatcher(conf)
	watcher = nil
	if err == nil || !strings.Contains(err.Error(), "error inotify instance already closed") {
		t.Fail()
	}
}

func TestAddWatcherWithNonExistingDir(t *testing.T) {
	tmpdir1 := "./tmp1"
	tmpdir2 := "./tmp2"
	users := []config.UserConf{
		{
			SourceDir: tmpdir1,
		},
		{
			SourceDir: tmpdir2,
		},
	}
	conf := config.Config{
		UsersConf: users,
	}
	err := AddWatcher(conf)
	if err == nil && strings.Contains(err.Error(), "Source directory ./tmp1 not found") {
		t.Fail()
	}
}

func TestWatchEvents(t *testing.T) {
	dir, _ := os.Getwd()
	tmpDir := "./tmp"
	users := []config.UserConf{
		{
			SourceDir: dir + "/tmp",
		},
	}
	conf := config.Config{
		UsersConf:       users,
		CleanupDuration: 60,
	}
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		os.MkdirAll(tmpDir+"/pm", os.ModePerm)
		os.MkdirAll(tmpDir+"/fm", os.ModePerm)
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	AddWatcher(conf)
	go WatchEvents(conf)
	defer os.RemoveAll("./tmp")

	tmpPMfile, _ := os.Create(tmpDir + "/pm/test_file.json")
	tmpPMfile.Write([]byte("test"))

	tmpFMfile, _ := os.Create(tmpDir + "/fm/test_file.json")
	tmpFMfile.Write([]byte("test"))
	time.Sleep(100 * time.Millisecond)
	if !strings.Contains(buf.String(), "Received event: "+dir+"/tmp/pm/test_file.json") || !strings.Contains(buf.String(), "Received event: "+dir+"/tmp/fm/test_file.json") {
		t.Fail()
	}
}

func TestProcessExistingFilesWithNonExistingDir(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	processExistingFiles("./tmp", config.Config{})
	if !strings.Contains(buf.String(), "./tmp: no such file or directory") {
		t.Fail()
	}

}

func TestProcessExistingFiles(t *testing.T) {
	testFMData := `[  
    {
       "_index":"test_index",
       "_type":"",
       "_id":"12345",
       "_source":{
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
          "EdgeName":""
       }
    },
    {
       "_index":"test_index",
       "_type":"",
       "_id":"12345",
       "_source":{
          "Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
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
          "ProbableCause":"Time synchronization failed",
          "AlarmText":"Time synchronization failed",
          "EventTime":"2018-09-24T16:28:01+05:30",
          "NotificationType":"ClearedAlarm",
          "AlarmIdentifier":"11191",
          "EdgeName":""
       }
    }
 ]`

	testPMData := `[
	{
	  "_index": "test_index",
	  "_id": "12345",
	  "_source": {
		"Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
		"edgename": "test edge",
		"SIG_SctpCongestionDuration": 0,
		"neserialno": "4355",
		"nename": "test",
		"edge_id": "edge_id",
		"SIG_SctpUnavailableDuration": 0,
		"edge_hostname": "localhost",
		"timestamp": "2018-11-22T16:15:00+05:30"
	  }
	},
	{
	  "_index": "test_index",
	  "_id": "124576",
	  "_source": {
		"Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
		"edgename": "test edge 2",
		"EQPT_MaxMeLoad": 0.5,
		"neserialno": "3456",
		"EQPT_MeanCpuUsage_0": 5,
		"EQPT_MeanCpuUsage_1": 5,
		"EQPT_MaxCpuUsage_1": 99,
		"edge_id": "edge_id2",
		"EQPT_MeanMeLoad": 0.4,
		"edge_hostname": "localhost",
		"ObjectType": "ManagedElement",
		"UserLabel": "temp",
		"nename": "476486",
		"EQPT_MaxCpuUsage_0": 100,
		"EQPT_MaxMemUsage": 59,
		"EQPT_MeanMemUsage": 59,
		"timestamp": "2018-11-22T16:15:00+05:30"
	  }
	}
  ]`
	pmDirPath := "./pmdata"
	os.MkdirAll(pmDirPath, os.ModePerm)
	defer os.RemoveAll(pmDirPath)

	fmDirPath := "./fmdata"
	os.MkdirAll(fmDirPath, os.ModePerm)
	defer os.RemoveAll(fmDirPath)

	err := createTestData("./fmdata/fm_data.json", testFMData)
	if err != nil {
		t.Error(err)
	}
	err = createTestData("./pmdata/pm_data.json", testPMData)
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	processExistingFiles(pmDirPath, config.Config{CleanupDuration: 60})
	if !strings.Contains(buf.String(), "Formatted PM data written") {
		t.Fail()
	}

	processExistingFiles(fmDirPath, config.Config{CleanupDuration: 60})
	if !strings.Contains(buf.String(), "Formatted FM data written") {
		t.Fail()
	}
}

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
