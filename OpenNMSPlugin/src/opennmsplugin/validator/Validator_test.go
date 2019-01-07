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
	tmpfile, err := ioutil.TempFile(".", "xml")
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
	err = ValidateXML(tmpfile.Name())
	if err != nil {
		t.Fail()
	}
}

func TestValidateXMLWithWrongFilePath(t *testing.T) {
	fileName := "./test_file.xml"
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
	tmpfile, err := ioutil.TempFile(".", "xml")
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
	err = ValidateXML(tmpfile.Name())
	if err == nil {
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
	ValidateDestDir(conf)
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
	ValidateDestDir(conf)
	defer os.RemoveAll(tmpDir1)
	defer os.RemoveAll(tmpDir2)
	if _, err := os.Stat(tmpDir1); os.IsNotExist(err) {
		t.Fail()
	}
	if _, err := os.Stat(tmpDir2); os.IsNotExist(err) {
		t.Fail()
	}
}
