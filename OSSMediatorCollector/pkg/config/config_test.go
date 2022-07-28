/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package config

import (
	"io/ioutil"
	"os"
	"testing"
)

//Reading config from invalid json file
func TestReadConfigWithEmptyFile(t *testing.T) {
	tmpfile, err := createTmpFile(".", "conf", []byte(``))
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile)
	err = ReadConfig(tmpfile)
	if err == nil {
		t.Fail()
	}
}

//Reading valid config
func TestReadConfig(t *testing.T) {
	content := []byte(`{
		"base_url": "https://localhost:8080/api/v2",
		"users":
		[
			{
				"email_id": "user1@nokia.com",
				"response_dest": "/statistics/reports/user1"
			},
			{
				"email_id": "user2@nokia.com",
				"response_dest": "/statistics/reports/user2"
			}
		],
		"um_api":
		{
			"login": "/session",
			"refresh": "/refresh",
			"logout": "/logout"
		},
		"sim_apis":
		[
			{
				"api": "/ndac/sims",
				"interval": 15
			}
		],
		"metric_apis":
		[
			{
				"api": "/ndac/pmdata",
				"interval": 15
			},
			{
				"api": "/ndac/fmdata",
				"type": "ACTIVE",
				"metric_type": "RADIO",
				"interval": 15
			}
		],
		"limit": 100
	}`)
	tmpfile, err := createTmpFile(".", "conf", content)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile)

	err = ReadConfig(tmpfile)
	if err != nil {
		t.Error(err)
	}
	if Conf.BaseURL != "https://localhost:8080/api/v2" || len(Conf.Users) != 2 || Conf.Users[0].Email != "user1@nokia.com" {
		t.Fail()
	}
}

//Reading config from non existing file
func TestReadConfigWithNonExistingFile(t *testing.T) {
	err := ReadConfig("conf.json")
	if err == nil {
		t.Fail()
	}
}

func createTmpFile(dir string, prefix string, content []byte) (string, error) {
	tmpfile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return "", err
	}
	if _, err = tmpfile.Write(content); err != nil {
		return "", err
	}
	if err = tmpfile.Close(); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}
