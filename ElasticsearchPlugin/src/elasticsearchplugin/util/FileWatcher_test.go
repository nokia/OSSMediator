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

	"elasticsearchplugin/config"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
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
	if err == nil && strings.Contains(err.Error(), "Source directory path can't be empty") {
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
	if err == nil && strings.Contains(err.Error(), "Source directory ./tmp1 not found") {
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
	defer os.RemoveAll(pmDirPath)

	fmDirPath := "./fmdata"
	os.MkdirAll(fmDirPath, os.ModePerm)
	defer os.RemoveAll(fmDirPath)

	fmFile := "./fmdata/fmdata_RADIO_ACTIVE_data.json"
	err := createTestData(fmFile, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fmFile)
	pmFile := "./pmdata/pmdata_data.json"
	err = createTestData(pmFile, testPMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(pmFile)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	conf := config.Config{
		CleanupDuration:  60,
		ElasticsearchURL: "http://127.0.0.1:9299",
	}

	processExistingFiles(pmDirPath, conf)
	if !strings.Contains(buf.String(), "Data from "+pmFile+" pushed to elasticsearch successfully") {
		t.Fail()
	}

	processExistingFiles(fmDirPath, conf)
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
