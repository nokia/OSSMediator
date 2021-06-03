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
	content := []byte(``)
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

	_, err = ReadConfig(tmpfile.Name())
	if err == nil {
		t.Fail()
	}
}

//Reading valid config
func TestReadConfig(t *testing.T) {
	content := []byte(`{
		"source_dirs": [
			"/statistics/report/brlsyveview1   ", "/statistics/report/brlsyveview2  "
		],
		"elasticsearch": {
			"url": "http://127.0.0.1:9200  ",
			"user": "user",
			"password": "dGVzdDE="
		}
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

	conf, err := ReadConfig(tmpfile.Name())
	if err != nil || conf.ElasticsearchConf.URL != "http://127.0.0.1:9200" || conf.ElasticsearchConf.Password != "test1" {
		t.Fail()
	}
}

//Reading config from non existing file
func TestReadConfigWithNonExistingFile(t *testing.T) {
	_, err := ReadConfig("conf.json")
	if err == nil {
		t.Fail()
	}
}
