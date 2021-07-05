/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestRemoveFiles(t *testing.T) {
	cleanupDuration := 1
	path := "./tmp"

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		t.Fail()
	}
	ioutil.TempFile(path, "test_file1")
	ioutil.TempFile(path, "test_file2")

	CleanUp(cleanupDuration, path)
	time.Sleep(61 * time.Second)
	defer os.RemoveAll(path)
	files, _ := ioutil.ReadDir(path)
	if len(files) != 0 {
		t.Fail()
	}
}

func TestRemoveFilesWithNonExistingDir(t *testing.T) {
	cleanupDuration := 1
	path := "./tmp"
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	RemoveFiles(cleanupDuration, path)
	if !strings.Contains(buf.String(), "no such file or directory") {
		t.Fail()
	}
}
