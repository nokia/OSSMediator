/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPlugin(t *testing.T) {
	content := []byte(`{
		"source_dirs": [
			"./statistics/reports/nmsviews1", "./statistics/reports/nmsviews2"
		],
		"elasticsearch": {
			"url": "http://127.0.0.1:9299"
		},
		"cleanup_duration": 60
	}`)
	tmpfile, err := ioutil.TempFile(".", "conf")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmpfile.Name())
	if _, err = tmpfile.Write(content); err != nil {
		t.Error(err)
	}
	if err = tmpfile.Close(); err != nil {
		t.Error(err)
	}
	os.MkdirAll("./statistics/reports/nmsviews1/pm", os.ModePerm)
	os.MkdirAll("./statistics/reports/nmsviews1/fm", os.ModePerm)
	os.MkdirAll("./statistics/reports/nmsviews2/pm", os.ModePerm)
	os.MkdirAll("./statistics/reports/nmsviews2/fm", os.ModePerm)
	defer os.RemoveAll("./log")
	defer os.RemoveAll("./statistics")
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{
		"cmd",
		"-log_dir", "./log",
		"-conf_file", tmpfile.Name(),
	}

	go main()
	time.Sleep(2 * time.Second)

	if logDir != "./log" || confFile != tmpfile.Name() {
		t.Fail()
	}
	logFile := logDir + "/ElasticsearchPlugin.log"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error(err)
	}
	logs, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(logs))

	if !strings.Contains(string(logs), "Added watcher to ./statistics/reports/nmsviews2/pm") || !strings.Contains(string(logs), "Added watcher to ./statistics/reports/nmsviews2/fm") || !strings.Contains(string(logs), "Added watcher to ./statistics/reports/nmsviews1/pm") || !strings.Contains(string(logs), "Added watcher to ./statistics/reports/nmsviews1/fm") {
		t.Fail()
	}
}

func TestInitLogger(t *testing.T) {
	logDir = "./log"
	initLogger(logDir, 5)
	logFile := logDir + "/ElasticsearchPlugin.log"
	defer os.RemoveAll(logDir)
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fail()
	}
}
