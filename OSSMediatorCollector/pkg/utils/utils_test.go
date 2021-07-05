/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package utils

import (
	"collector/pkg/config"
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
	if endTime != "2021-06-17T19:59:50Z" {
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
	if endTime != "2021-06-17T19:59:50Z" {
		t.Fail()
	}
}

func TestCallAPIWIthLastReceivedFile(t *testing.T) {
	user := &config.User{Email: "testuser@nokia.com"}
	api := &config.APIConf{API: "/fmdata", Interval: 15, SyncDuration: 30}
	nhgID := "test_nhg_1"
	lastDataTime := "2018-03-14T13:55:00+05:30"

	tmpFile := "fmdata" + "_" + user.Email + "_" + nhgID
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	_, err = file.Write([]byte(lastDataTime))
	if err != nil {
		t.Error(err)
	}

	startTime, _ := GetTimeInterval(user, api, nhgID)
	if startTime != lastDataTime {
		t.Fail()
	}
	t.Log(os.Remove(tmpFile))
}
