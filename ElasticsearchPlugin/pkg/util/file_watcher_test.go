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

	"elasticsearchplugin/pkg/config"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var (
	testPMData = `[
    {
      "pm_data": {
        "Cat_M_Accessibility_M8100C0": 0,
        "Cat_M_Accessibility_M8100C10": 0,
        "Cat_M_Accessibility_M8100C11": 0
      },
      "pm_data_source": {
        "edge_id": "test_edge",
        "hw_id": "EB34567",
        "hw_alias": "EB34567",
        "serial_no": "34567",
        "nhg_id": "test_nhg_1",
        "nhg_alias": "test nhg",
        "dn": "NE-MRBTS-111/NE-LNBTS-222/LNCEL-0",
        "timestamp": "2020-11-10T18:30:00Z",
        "technology": "4G"
      }
    },
    {
      "pm_data": {
        "Cat_M_Accessibility_M8100C0": 0,
        "Cat_M_Accessibility_M8100C10": 0,
        "Cat_M_Accessibility_M8100C11": 0
      },
      "pm_data_source": {
        "edge_id": "test_edge2",
        "hw_id": "EB12345",
        "hw_alias": "EB12345",
        "serial_no": "12345",
        "nhg_id": "test_nhg_2",
        "nhg_alias": "test nhg2",
        "dn": "MRBTS-222/NE-LNBTS-4444/LNCEL-0",
        "timestamp": "2020-11-10T18:30:00Z",
        "technology": "4G"
      }
    }
  ]`

	testFMData = `[
    {
      "fm_data": {
        "alarm_identifier": "11111",
        "severity": "minor",
        "specific_problem": "2222",
        "alarm_text": "Failure in connection",
        "additional_text": "TraceConnectionFaultyAl",
        "alarm_state": "CLEARED",
        "event_type": "equipment",
        "event_time": "2020-11-02T06:14:09Z",
        "last_updated_time": "2020-11-02T06:14:10Z",
        "notification_type": "alarmNew"
      },
      "fm_data_source": {
        "edge_id": "98765",
        "hw_id": "EB09856",
        "hw_alias": "eNB7777",
        "serial_no": "45678",
        "nhg_id": "test_nhg_2",
        "nhg_alias": "test nhg2",
        "dn": "MRBTS-456/LNBTS-765",
        "technology": "4G"
      }
    },
    {
      "fm_data": {
        "alarm_identifier": "3333",
        "severity": "minor",
        "specific_problem": "4444",
        "alarm_state": "CLEARED",
        "event_type": "equipment",
        "event_time": "2020-11-02T06:14:10Z",
        "last_updated_time": "2020-11-02T06:14:10Z",
        "notification_type": "alarmClear"
      },
      "fm_data_source": {
        "edge_id": "test_edge",
        "hw_id": "EB1111111",
        "hw_alias": "test eNB",
        "serial_no": "123456",
        "nhg_id": "test_nhg_1",
        "nhg_alias": "test nhg",
        "dn": "MRBTS-123/LNBTS-321",
        "technology": "4G"
      }
    }
  ]`
)

func TestAddWatcher(t *testing.T) {
	tmpdir1 := "./tmp1"
	tmpdir2 := "./tmp2"
	conf := config.Config{
		SourceDirs:      []string{tmpdir1, tmpdir2},
		CleanupDuration: 60,
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
	conf := config.Config{
		SourceDirs:      []string{""},
		CleanupDuration: 60,
	}
	err := AddWatcher(conf)
	if err == nil && strings.Contains(err.Error(), "source directory path can't be empty") {
		t.Fail()
	}
}

func TestAddWatcherWithClosedWatcherInstance(t *testing.T) {
	tmpdir := "./tmp"
	conf := config.Config{
		SourceDirs:      []string{tmpdir},
		CleanupDuration: 60,
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
	conf := config.Config{
		SourceDirs:      []string{tmpdir1, tmpdir2},
		CleanupDuration: 60,
	}
	err := AddWatcher(conf)
	if err == nil && strings.Contains(err.Error(), "source directory ./tmp1 not found") {
		t.Fail()
	}
}

func TestWatchEvents(t *testing.T) {
	dir, _ := os.Getwd()
	tmpDir := "./tmp"
	conf := config.Config{
		SourceDirs:      []string{dir + "/tmp"},
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
	time.Sleep(10 * time.Millisecond)
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
	processExistingFiles("./tmp", config.Config{CleanupDuration: 60})
	if !strings.Contains(buf.String(), "./tmp: no such file or directory") {
		t.Fail()
	}

}

func TestProcessExistingFiles(t *testing.T) {
	pmDirPath := "./pmdata"
	os.MkdirAll(pmDirPath, os.ModePerm)

	fmDirPath := "./fmdata"
	os.MkdirAll(fmDirPath, os.ModePerm)

	fmFile := "./fmdata/fmdata_RADIO_ACTIVE_data.json"
	err := createTestData(fmFile, testFMData)
	if err != nil {
		t.Error(err)
	}
	pmFile := "./pmdata/pmdata_data.json"
	err = createTestData(pmFile, testPMData)
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	conf := config.Config{
		CleanupDuration: 60,
		ElasticsearchConf: config.ElasticsearchConf{
			URL: "http://127.0.0.1:9299",
		},
	}

	processExistingFiles(pmDirPath, conf)
	processExistingFiles(fmDirPath, conf)
	time.Sleep(200 * time.Millisecond)
	os.RemoveAll(fmDirPath)
	os.RemoveAll(pmDirPath)

	if !strings.Contains(buf.String(), "Data from "+pmFile+" pushed to elasticsearch successfully") {
		t.Fail()
	}

	if !strings.Contains(buf.String(), "Data from "+fmFile+" pushed to elasticsearch successfully") {
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
