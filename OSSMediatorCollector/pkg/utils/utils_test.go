/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package utils

import (
	"collector/pkg/config"
	"encoding/base64"
	"os"
	"testing"
	"time"
)

func TestGetTimeInterval(t *testing.T) {
	oldCurrentTime := CurrentTime
	defer func() { CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2021, 06, 17, 20, 9, 0, 0, time.UTC)
	}
	CurrentTime = myCurrentTime
	user := &config.User{Email: "testuser@nokia.com"}
	api := &config.APIConf{API: "/fmdata", Interval: 15}
	nhgID := "test_nhg_1"

	startTime, endTime := GetTimeInterval(user, api, nhgID)
	if startTime != "2021-06-17T19:45:00Z" {
		t.Fail()
	}
	if endTime != "2021-06-17T20:00:00Z" {
		t.Fail()
	}
}

func TestGetTimeIntervalWithSyncDuration(t *testing.T) {
	oldCurrentTime := CurrentTime
	defer func() { CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2021, 06, 17, 20, 9, 0, 0, time.UTC)
	}
	CurrentTime = myCurrentTime
	user := &config.User{Email: "testuser@nokia.com"}
	api := &config.APIConf{API: "/fmdata", Interval: 15, SyncDuration: 30}
	nhgID := "test_nhg_1"

	startTime, endTime := GetTimeInterval(user, api, nhgID)
	if startTime != "2021-06-17T19:30:00Z" {
		t.Fail()
	}
	if endTime != "2021-06-17T20:00:00Z" {
		t.Fail()
	}
}

func TestCallAPIWIthLastReceivedFile(t *testing.T) {
	user := &config.User{Email: "testuser@nokia.com"}
	api := &config.APIConf{API: "/fmdata", Interval: 15, SyncDuration: 30}
	nhgID := "test_nhg_1"
	lastDataTime := "2018-03-14T13:55:00+05:30"

	tmpFile := "checkpoints/fmdata" + "_" + user.Email + "_" + nhgID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	defer os.Remove(tmpFile)
	_, err = file.Write([]byte(lastDataTime))
	if err != nil {
		t.Error(err)
	}

	startTime, _ := GetTimeInterval(user, api, nhgID)
	if startTime != lastDataTime {
		t.Fail()
	}
}

func TestReadSessionToken(t *testing.T) {
	secretDir := ".secret"
	err := os.Mkdir(secretDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	emailID := "testuser@nokia.com"
	tmpFile := secretDir + "/." + emailID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	defer os.RemoveAll(secretDir)

	encodedPassword := base64.StdEncoding.EncodeToString([]byte("test")) + "\n" + base64.StdEncoding.EncodeToString([]byte("test2"))
	_, err = file.Write([]byte(encodedPassword))
	if err != nil {
		t.Error(err)
	}

	password, err := ReadSessionToken(emailID)
	if password != "test\ntest2" {
		t.Fail()
	}
	if err != nil {
		t.Fail()
	}
}

func TestReadPassword(t *testing.T) {
	secretDir := ".secret"
	err := os.Mkdir(secretDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	emailID := "testuser@nokia.com"
	tmpFile := secretDir + "/." + emailID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	defer os.RemoveAll(secretDir)

	encodedPassword := base64.StdEncoding.EncodeToString([]byte("test"))
	_, err = file.Write([]byte(encodedPassword))
	if err != nil {
		t.Error(err)
	}

	password, err := ReadPassword(emailID)
	if password != "test" {
		t.Fail()
	}
	if err != nil {
		t.Fail()
	}
}

func TestReadPasswordWithInvalidFile(t *testing.T) {
	password, err := ReadPassword("testuser@nokia.com")
	if password != "" {
		t.Fail()
	}
	if err != errorPasswordFileNotFound {
		t.Fail()
	}
}

func TestReadSessionTokenWithInvalidFile(t *testing.T) {
	password, err := ReadSessionToken("testuser@nokia.com")
	if password != "" {
		t.Fail()
	}
	if err != errorPasswordFileNotFound {
		t.Fail()
	}
}

func TestReadPasswordWithEmptyFile(t *testing.T) {
	secretDir := ".secret"
	err := os.Mkdir(secretDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	emailID := "testuser@nokia.com"
	tmpFile := secretDir + "/." + emailID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	defer os.RemoveAll(secretDir)

	password, err := ReadPassword(emailID)
	if password != "" {
		t.Fail()
	}
	if err != errorPasswordRead {
		t.Fail()
	}
}

func TestReadSessionTokenWithEmptyFile(t *testing.T) {
	secretDir := ".secret"
	err := os.Mkdir(secretDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	emailID := "testuser@nokia.com"
	tmpFile := secretDir + "/." + emailID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	defer os.RemoveAll(secretDir)

	password, err := ReadSessionToken(emailID)
	if password != "" {
		t.Fail()
	}
	if err != errorPasswordRead {
		t.Fail()
	}
}

func TestReadPasswordWithInvalidEncoding(t *testing.T) {
	secretDir := ".secret"
	err := os.Mkdir(secretDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	emailID := "testuser@nokia.com"
	tmpFile := secretDir + "/." + emailID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	defer os.RemoveAll(secretDir)

	_, err = file.Write([]byte("???"))
	if err != nil {
		t.Error(err)
	}

	password, err := ReadPassword(emailID)
	if password != "" {
		t.Fail()
	}
	if err != errorPasswordDecoding {
		t.Fail()
	}
}

func TestReadSessionTokenWithInvalidEncoding(t *testing.T) {
	secretDir := ".secret"
	err := os.Mkdir(secretDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	emailID := "testuser@nokia.com"
	tmpFile := secretDir + "/." + emailID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	defer os.RemoveAll(secretDir)

	_, err = file.Write([]byte("???"))
	if err != nil {
		t.Error(err)
	}

	password, err := ReadSessionToken(emailID)
	if password != "" {
		t.Fail()
	}
	if err != errorPasswordDecoding {
		t.Fail()
	}
}
