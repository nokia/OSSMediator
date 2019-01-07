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
		UsersConf: users,
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

	tmpPMfile, _ := os.Create(tmpDir + "/pm/test_file.xml")
	tmpPMfile.Write([]byte("test"))

	tmpFMfile, _ := os.Create(tmpDir + "/fm/test_file.csv")
	tmpFMfile.Write([]byte("test"))
	time.Sleep(100 * time.Millisecond)
	if !strings.Contains(buf.String(), "Received event: "+dir+"/tmp/pm/test_file.xml") || !strings.Contains(buf.String(), "Received event: "+dir+"/tmp/fm/test_file.csv") {
		t.Error(buf.String())
		t.Fail()
	}
}
