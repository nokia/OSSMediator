/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */
 
 package validator

import (
	"io/ioutil"
	"opennmsplugin/config"
	"os"
	"strings"
	"testing"
)

func TestValidateXML(t *testing.T) {
	content := []byte(`
		<?xml version="1.0" encoding="UTF-8"?>
		<a>
			<b>test</b>
		</a>
	`)
	tmpfile, err := createTempFile(".", "xml", content)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmpfile)
	err = ValidateXML(tmpfile)
	if err != nil {
		t.Error(err)
	}
}

func TestValidateXMLWithWrongFilePath(t *testing.T) {
	fileName := "./file.xml"
	err := ValidateXML(fileName)
	if err == nil || !strings.Contains(err.Error(), "no such file or directory") {
		t.Fail()
	}
}

func TestValidateXMLWithInvalid(t *testing.T) {
	content := []byte(`
		<?xml version="1.0" encoding="UTF-8"?>
		<a>
			<b>test</b>
	`)
	tmpfile, err := createTempFile(".", "xml", content)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmpfile)
	err = ValidateXML(tmpfile)
	if err == nil && !strings.Contains(err.Error(), "unexpected EOF") {
		t.Fail()
	}
}

func TestValidateDestDir(t *testing.T) {
	tmpDir1 := "./tmp1"
	tmpDir2 := "./tmp2"
	users := []config.UserConf{
		{
			PMConfig: config.PMConfig{
				DestinationDir: tmpDir1,
			},
			FMConfig: config.FMConfig{
				DestinationDir: tmpDir2,
			},
		},
	}
	conf := config.Config{
		UsersConf: users,
	}
	err := os.MkdirAll(tmpDir1, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	err = os.MkdirAll(tmpDir2, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	CreateDestDir(conf)
	defer os.RemoveAll(tmpDir1)
	defer os.RemoveAll(tmpDir2)
	if _, err := os.Stat(tmpDir1); os.IsNotExist(err) {
		t.Fail()
	}
	if _, err := os.Stat(tmpDir2); os.IsNotExist(err) {
		t.Fail()
	}
}

func TestValidateDestDirWithNonExistingDirectory(t *testing.T) {
	tmpDir1 := "./tmp1"
	tmpDir2 := "./tmp2"
	users := []config.UserConf{
		{
			PMConfig: config.PMConfig{
				DestinationDir: tmpDir1,
			},
			FMConfig: config.FMConfig{
				DestinationDir: tmpDir2,
			},
		},
	}
	conf := config.Config{
		UsersConf: users,
	}
	CreateDestDir(conf)
	defer os.RemoveAll(tmpDir1)
	defer os.RemoveAll(tmpDir2)
	if _, err := os.Stat(tmpDir1); os.IsNotExist(err) {
		t.Fail()
	}
	if _, err := os.Stat(tmpDir2); os.IsNotExist(err) {
		t.Fail()
	}
}

func createTempFile(path string, prefix string, content []byte) (string, error) {
	tmpfile, err := ioutil.TempFile(path, prefix)
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

func TestValidateJSON(t *testing.T) {
	content := []byte(`
	{
		"colors": [
		  {
			"color": "red",
			"category": "hue",
			"type": "primary",
			"code": {
			  "rgba": [255,0,0,1],
			  "hex": "#FF0"
			}
		  },
		  {
			"color": "blue",
			"category": "hue",
			"type": "primary",
			"code": {
			  "rgba": [0,0,255,1],
			  "hex": "#00F"
			}
		  }
		]
	  }
	`)
	tmpfile, err := createTempFile(".", "json", content)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmpfile)
	err = ValidateJSON(tmpfile)
	if err != nil {
		t.Fail()
	}
}

func TestValidateJSONWithWrongFilePath(t *testing.T) {
	fileName := "./file.json"
	err := ValidateJSON(fileName)
	if err == nil || !strings.Contains(err.Error(), "no such file or directory") {
		t.Fail()
	}
}

func TestValidateJSONWithInvalid(t *testing.T) {
	content := []byte(`
	{
		"colors": [
		  {}
		]
	`)
	tmpfile, err := createTempFile(".", "json", content)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmpfile)
	err = ValidateJSON(tmpfile)
	if err == nil && !strings.Contains(err.Error(), "unexpected end of JSON input") {
		t.Error(err)
	}
}
