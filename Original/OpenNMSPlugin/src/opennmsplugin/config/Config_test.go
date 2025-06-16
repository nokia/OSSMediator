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
		"users_conf": [
			{
				"source_dir": "/statistics/reports/nmsviews1",
				"pm_config": {
					"destination_dir": "/statistics/reports/nmsviews1_pm",
					"foreign_id": "1234567"
				},
				"fm_config": {
					"source": "NDAC",
					"node_id": "1",
					"host": "127.0.0.1",
					"service": "NDAC",
					"destination_dir": "/statistics/reports/nmsviews1_fm"
				}
			},
			{
				"source_dir": "/statistics/reports/nmsviews2",
				"pm_config": {
					"destination_dir": "/statistics/reports/nmsviews2_pm",
					"foreign_id": "9876543"
				},
				"fm_config": {
					"source": "NDAC",
					"node_id": "1",
					"host": "127.0.0.1",
					"service": "NDAC",
					"destination_dir": "/statistics/reports/nmsviews2_fm"
				}
			}
		],
		"opennms_address": "127.0.0.1:5817",
		"cleanup_duration": 60
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
	if err != nil || conf.OpenNMSAddress != "127.0.0.1:5817" {
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
