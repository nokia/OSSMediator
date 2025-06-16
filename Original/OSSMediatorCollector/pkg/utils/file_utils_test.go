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
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_Success(t *testing.T) {
	fileName := "testfile.txt"
	data := []byte("This is a test")

	// Call the function under test
	err := writeFile(fileName, data)

	// Assert that no error occurred
	assert.Nil(t, err)

	// Read the file and verify its content
	fileContent, err := ioutil.ReadFile(fileName)
	assert.Nil(t, err)
	assert.Equal(t, data, fileContent)

	// Clean up: remove the created file
	err = os.Remove(fileName)
	assert.Nil(t, err)
}

func TestWriteFile_FileCreationError(t *testing.T) {
	fileName := "/nonexistentfolder/testfile.txt"
	data := []byte("This is a test")

	// Call the function under test
	err := writeFile(fileName, data)

	// Assert that an error occurred during file creation
	expectedError := fmt.Errorf("file creation failed")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), expectedError.Error())
}

func TestWriteResponseWithWrongDir(t *testing.T) {
	user := &config.User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	apiConf := &config.APIConf{API: "/fmdata", Type: "ACTIVE", MetricType: "RADIO"}
	err := WriteResponse(user, apiConf, "", "", 123, true)
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
		err := WriteResponse(user, &api, data, "", 123, true)
		if err != nil {
			t.Error(err)
		}

		files, err := ioutil.ReadDir(user.ResponseDest + api.API)

		if err != nil {
			t.Error(err)
		}
		content, err := ioutil.ReadFile(user.ResponseDest + api.API + "/" + files[0].Name())

		if err != nil || len(content) == 0 {
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

func TestCreateResponseDirectory_Success(t *testing.T) {
	basePath := "/path/to/base"
	api := "/api/v1/foo"

	// Call the function under test
	CreateResponseDirectory(basePath, api)

	// Assert that the directory exists
	dirPath := filepath.Join(basePath, filepath.Base(api))
	_, err := os.Stat(dirPath)
	assert.Nil(t, err)

	// Clean up: remove the created directory
	err = os.RemoveAll(dirPath)
	assert.Nil(t, err)
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
	_ = os.Mkdir("checkpoints", os.ModePerm)

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
