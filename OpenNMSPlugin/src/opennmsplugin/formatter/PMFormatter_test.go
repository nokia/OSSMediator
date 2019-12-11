/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package formatter

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"opennmsplugin/config"

	log "github.com/sirupsen/logrus"
)

var testPMData = `[
	{
	  "_index": "test_index",
	  "_id": "12345",
	  "_source": {
		"Dn": "NE-MRBTS-1/NE-LNBTS-1/LNCEL-12/MCC-244/MNC-09",
		"edgename": "test edge",
		"SIG_SctpCongestionDuration": 0,
		"neserialno": "4355",
		"nename": "test",
		"edge_id": "edge_id",
		"edge_hostname": "localhost",
		"timestamp": "2018-11-22T16:15:00+05:30"
	  }
	},
	{
	  "_index": "test_index",
	  "_id": "124576",
	  "_source": {
		"Dn": "NE-MRBTS-1/NE-LNBTS-1/MCC-244/MNC-09",
		"edgename": "test edge 2",
		"EQPT_MaxMeLoad": 0.5,
		"neserialno": "3456",
		"EQPT_MeanCpuUsage_0": 5,
		"EQPT_MeanCpuUsage_1": 5,
		"EQPT_MaxCpuUsage_1": 99,
		"edge_id": "edge_id2",
		"EQPT_MeanMeLoad": 0.4,
		"edge_hostname": "localhost",
		"ObjectType": "ManagedElement",
		"UserLabel": "temp",
		"nename": "476486",
		"EQPT_MaxCpuUsage_0": 100,
		"EQPT_MaxMemUsage": 59,
		"EQPT_MeanMemUsage": 59,
		"timestamp": "2018-11-22T16:15:00+05:30"
	  }
	}
  ]`

func TestRenameFile(t *testing.T) {
	pmConfig := config.PMConfig{
		DestinationDir: "./tmp",
		ForeignID:      "12345",
	}
	fileName, err := renameFile(pmConfig)
	if fileName == "" && err != nil {
		t.Fail()
	}
}

func TestFormatPMData(t *testing.T) {
	pmConfig := config.PMConfig{
		DestinationDir: "./tmp",
		ForeignID:      "12345",
	}
	fileName := "./pm_data.json"
	err := createTestData(fileName, testPMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	err = os.MkdirAll(pmConfig.DestinationDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(pmConfig.DestinationDir)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	FormatPMData(fileName, pmConfig)
	if !strings.Contains(buf.String(), "Formatted PM data written") {
		t.Fail()
	}
}

func TestFormatPMDataWithNonExistingDir(t *testing.T) {
	pmConfig := config.PMConfig{
		DestinationDir: "./tmp",
		ForeignID:      "12345",
	}
	fileName := "./pm_data.json"
	err := createTestData(fileName, testPMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	FormatPMData(fileName, pmConfig)
	if !strings.Contains(buf.String(), "no such file or directory") {
		t.Fail()
	}
}

func TestFormatPMDataWithNonExistingFile(t *testing.T) {
	pmConfig := config.PMConfig{
		DestinationDir: "./tmp",
		ForeignID:      "12345",
	}

	err := os.MkdirAll(pmConfig.DestinationDir, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(pmConfig.DestinationDir)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	FormatPMData("test_file.json", pmConfig)
	if !strings.Contains(buf.String(), "no such file or directory") {
		t.Fail()
	}
}
