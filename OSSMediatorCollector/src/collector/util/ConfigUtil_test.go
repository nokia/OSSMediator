/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

//Reading config from invalid json file
func TestReadConfigWithEmptyFile(t *testing.T) {
	tmpfile := createTmpFile(".", "conf", []byte(``))
	defer os.Remove(tmpfile)
	err := ReadConfig(tmpfile)
	if err == nil {
		t.Fail()
	}
}

//Reading valid config
func TestReadConfig(t *testing.T) {
	content := []byte(`{
		"base_url": "https://localhost:8080",
		"users":
		[
			{
				"email_id" : "user1@nokia.com",
				"response_dest": "/statistics/reports/pm_source_1"
			},
			{
				"email_id" : "user2@nokia.com",
				"response_dest": "/statistics/reports/pm_source_2"
			}
		],
		"um_api":
		{
			"login": "/session",
			"refresh": "/refresh",
			"logout": "/logout"
		},
		"apis":
		[
			{
				"api": "/nms/pmdata",
				"interval": 15
			}
		]
	}`)
	tmpfile := createTmpFile(".", "conf", content)
	defer os.Remove(tmpfile)

	oldReadPassword := readPassword
	defer func() { readPassword = oldReadPassword }()

	myReadPassword := func(int) ([]byte, error) {
		return []byte("password"), nil
	}
	readPassword = myReadPassword
	err := ReadConfig(tmpfile)
	if err != nil {
		t.Error(err)
	}
	if Conf.BaseURL != "https://localhost:8080" || len(Conf.Users) != 2 || Conf.Users[0].Email != "user1@nokia.com" {
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

func TestReadCredentials(t *testing.T) {
	oldReadPassword := readPassword
	defer func() { readPassword = oldReadPassword }()

	myReadPassword := func(int) ([]byte, error) {
		return []byte("password"), nil
	}
	readPassword = myReadPassword
	Conf.Users = []*User{&User{Email: "user@nokia.com"}}
	ReadCredentials()
	if Conf.Users[0].Email != "user@nokia.com" && Conf.Users[0].password != "password" {
		t.Fail()
	}
}

func createTmpFile(dir string, prefix string, content []byte) string {
	tmpfile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		log.Error(err)
		return ""
	}

	if _, err = tmpfile.Write(content); err != nil {
		log.Error(err)
		return ""
	}
	if err = tmpfile.Close(); err != nil {
		log.Error(err)
		return ""
	}
	return tmpfile.Name()
}
