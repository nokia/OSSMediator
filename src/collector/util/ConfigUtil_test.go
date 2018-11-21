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
)

//Reading config from invalid json file
func TestReadConfigWithEmptyFile(t *testing.T) {
	content := []byte(``)
	tmpfile, err := ioutil.TempFile(".", "conf")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	defer os.Remove(tmpfile.Name())
	if _, err = tmpfile.Write(content); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err = tmpfile.Close(); err != nil {
		t.Log(err)
		t.Fail()
	}

	err = ReadConfig(tmpfile.Name())
	if err == nil {
		t.Fail()
	}
}

//Reading valid config
func TestReadConfig(t *testing.T) {
	content := []byte(`{
		"base_url": "https://api.ci-dev2.daaas.dynamic.nsn-net.net/api/v2",
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
	tmpfile, err := ioutil.TempFile(".", "conf")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	defer os.Remove(tmpfile.Name())
	if _, err = tmpfile.Write(content); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err = tmpfile.Close(); err != nil {
		t.Log(err)
		t.Fail()
	}

	oldReadPassword := readPassword
	defer func() { readPassword = oldReadPassword }()

	myReadPassword := func(int) ([]byte, error) {
		return []byte("password"), nil
	}
	readPassword = myReadPassword
	err = ReadConfig(tmpfile.Name())
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if Conf.BaseURL != "https://api.ci-dev2.daaas.dynamic.nsn-net.net/api/v2" || len(Conf.Users) != 2 || Conf.Users[0].Email != "user1@nokia.com" {
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
	if Conf.Users[0].Email != "user@nokia.com" && Conf.Users[0].Email != "password" {
		t.Fail()
	}
}
