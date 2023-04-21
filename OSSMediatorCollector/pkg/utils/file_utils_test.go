/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package utils

import (
	"bytes"
	"collector/pkg/config"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestWriteResponseWithWrongDir(t *testing.T) {
	user := &config.User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	apiConf := &config.APIConf{API: "/fmdata", Type: "ACTIVE", MetricType: "RADIO"}
	err := WriteResponse(user, apiConf, "", "", 123)
	if err == nil {
		t.Error(err)
	}
}

func TestWriteResponseForPM(t *testing.T) {
	user := &config.User{Email: "testuser@nokia.com", ResponseDest: "./tmp"}
	apiConfs := []config.APIConf{
		{API: "/pmdata", Interval: 15},
		{API: "/fmdata", MetricType: "DAC", Type: "ACTIVE", Interval: 15},
		{API: "/sims", Interval: 15},
		{API: "/network-hardware-groups", Interval: 15},
	}
	data := "test"
	for _, api := range apiConfs {
		CreateResponseDirectory(user.ResponseDest, api.API)
	}
	defer os.RemoveAll(user.ResponseDest)

	for _, api := range apiConfs {
		err := WriteResponse(user, &api, data, "", 123)
		if err != nil {
			t.Error(err)
		}

		files, err := ioutil.ReadDir(user.ResponseDest + api.API)
		if err != nil {
			t.Error(err)
		}
		content, err := ioutil.ReadFile(user.ResponseDest + api.API + "/" + files[0].Name())
		if err != nil || len(content) == 0 || !strings.Contains(string(content), data) {
			t.Fail()
		}
	}
}

func TestCreateResponseDirectory(t *testing.T) {
	respDir := "./tmp"
	CreateResponseDirectory(respDir, "http://localhost:8080/pmdata")
	defer os.RemoveAll(respDir)
	if _, err := os.Stat(respDir + "/pmdata"); os.IsNotExist(err) {
		t.Fail()
	}
}

func TestStoreLastReceivedDataTimeForFM(t *testing.T) {
	var data []interface{}
	responseData := `[{"fm_data":{"event_time":"2020-10-30T13:29:27Z"}},{"fm_data":{"event_time":"2020-10-30T13:39:27Z"}},{"fm_data":{"event_time":"2020-10-30T13:29:29Z"}},{"fm_data":{"event_time":"2020-10-30T13:39:29Z"}}]`
	err := json.NewDecoder(bytes.NewReader([]byte(responseData))).Decode(&data)
	if err != nil {
		t.Error(err)
	}
	user := &config.User{Email: "testuser@nokia.com"}
	api := &config.APIConf{API: "/fmdata", Type: "ACTIVE", MetricType: "DAC"}
	nhgID := "test_nhg_1"
	err = os.Mkdir("checkpoints", os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	err = StoreLastReceivedDataTime(user, data, api, nhgID, 123)
	if err != nil {
		t.Error(err)
	}
	//Reading LastReceivedFile value from file
	fileName := "checkpoints/fmdata" + "_" + api.MetricType + "_" + api.Type + "_" + user.Email + "_" + nhgID
	content, err := ioutil.ReadFile(fileName)
	defer os.Remove(fileName)
	if err != nil && len(content) == 0 && string(content) != "2020-10-30T13:39:00Z" {
		t.Fail()
	}

	lastReceivedTime := getLastReceivedDataTime(user, api, nhgID)
	if lastReceivedTime != "2020-10-30T13:39:00Z" {
		t.Fail()
	}
}

func TestStoreLastReceivedDataTimeForPM(t *testing.T) {
	var data []interface{}
	responseData := `[{"pm_data_source":{"timestamp":"2020-10-30T13:29:27Z"}},{"pm_data_source":{"timestamp":"2020-10-30T13:39:27Z"}},{"pm_data_source":{"timestamp":"2020-10-30T13:29:29Z"}},{"pm_data_source":{"timestamp":"2020-10-30T13:39:29Z"}}]`
	err := json.NewDecoder(bytes.NewReader([]byte(responseData))).Decode(&data)
	if err != nil {
		t.Error(err)
	}
	user := &config.User{Email: "testuser@nokia.com"}
	api := &config.APIConf{API: "/pmdata"}
	nhgID := "test_nhg_1"
	err = StoreLastReceivedDataTime(user, data, api, nhgID, 123)
	if err != nil {
		t.Error(err)
	}
	//Reading LastReceivedFile value from file
	fileName := "checkpoints/pmdata" + "_" + user.Email + "_" + nhgID
	content, err := ioutil.ReadFile(fileName)
	defer os.Remove(fileName)
	if err != nil && len(content) == 0 && string(content) != "2020-10-30T13:39:00Z" {
		t.Fail()
	}

	lastReceivedTime := getLastReceivedDataTime(user, api, nhgID)
	if lastReceivedTime != "2020-10-30T13:39:00Z" {
		t.Fail()
	}
}
