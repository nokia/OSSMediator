/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
*/

package formatter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"opennmsplugin/config"
	"opennmsplugin/validator"

	log "github.com/sirupsen/logrus"
)

const (
	pmFileNameFormat = `PM(?P<file_create_date>\d{8})(?P<file_create_time>\d{4})(?P<offset>[\+-]\d{4})\d{0,}\w{0,}(?:_-_|)(?P<seq_no>\d{0,})\.xml`

	//regex captures
	offsetCapture         = "offset"
	fileCreateDateCapture = "file_create_date"
	fileCreateTimeCapture = "file_create_time"
	seqNoCapture          = "seq_no"

	//pmdata xml tags
	measInfo       = "measInfo"
	granPeriod     = "granPeriod"
	repPeriod      = "repPeriod"
	measData       = "measData"
	measCollecFile = "measCollecFile"
	measTypes      = "measTypes"
	measResults    = "measResults"
)

//FormatPMData formats PM Data from A2 to A3 format
func FormatPMData(filePath string, pmConfig config.PMConfig) {
	log.Infof("Formatting PM file %s", filePath)
	f, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while formatting pm data %s", filePath)
		return
	}
	defer f.Close()
	//removing source file
	defer os.Remove(filePath)
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	fileName := filepath.Base(f.Name())
	fileName = renameFile(fileName, pmConfig.ForeignID)
	if fileName == "" {
		return
	}
	fileName = pmConfig.DestinationDir + "/" + fileName
	opFile, err := os.Create(fileName)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while formatting pm data %s", filePath)
		return
	}
	defer opFile.Close()

	var line string
	for scanner.Scan() {
		line = scanner.Text()
		j := 1
		if strings.Contains(line, measInfo) {
			fmt.Fprintf(opFile, "%s\n", line)
			continue
		}
		if strings.Contains(line, granPeriod) {
			fmt.Fprintf(opFile, "%s\n", line)
			continue
		}
		if strings.Contains(line, repPeriod) {
			fmt.Fprintf(opFile, "%s\n", line)
			continue
		}
		if strings.Contains(line, "/"+measInfo) {
			fmt.Fprintf(opFile, "%s\n", line)
			line = scanner.Text()
			temp := string(line)
			if !strings.Contains(temp, "/"+measData) {
				continue
			} else {
				if j == 1 {
					fmt.Fprintf(opFile, "%s\n", line)
					fmt.Fprintf(opFile, "%s\n", temp)
					continue
				}
			}
		}
		if strings.Contains(line, measCollecFile+" ") {
			fmt.Fprintf(opFile, "%s\n", "<"+measCollecFile+">")
			continue
		}
		if strings.Contains(line, measTypes) {
			words := regexp.MustCompile("[ </>]").Split(line, -1)
			if words != nil {
				for _, word := range words {
					ifExists, _ := regexp.MatchString(word, measTypes)
					if !ifExists {
						fmt.Fprintf(opFile, "<measType p=\"%d\">%s</measType>\n", j, word)
						j++
					}
				}
			}
			continue
		}
		if strings.Contains(line, measResults) {
			words := regexp.MustCompile("[ </>]").Split(line, -1)
			if words != nil {
				for _, word := range words {
					ifExists, _ := regexp.MatchString(word, measResults)
					if !ifExists {
						fmt.Fprintf(opFile, "<r p=\"%d\">%s</r>\n", j, word)
						j++
					}
				}
			}
		}
		if j == 1 {
			fmt.Fprintf(opFile, "%s\n", line)
		}
	}

	//Validate the xml
	err = validator.ValidateXML(fileName)
	if err != nil {
		log.Errorf("Invalid xml generated from %s", filePath)
		os.Remove(fileName)
		return
	}
	log.Infof("Formatted PM data written to %s", fileName)
}

//Rename the file to OpenNMS format
func renameFile(fileName string, foreignID string) string {
	r := regexp.MustCompile(pmFileNameFormat)
	captures := make(map[string]string)
	match := r.FindStringSubmatch(fileName)
	if match == nil {
		log.Errorf("%s din't match with the file name format", fileName)
		return ""
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}
		captures[name] = match[i]
	}

	offset := captures[offsetCapture]
	offset = offset[:3] + "h" + offset[3:] + "m"
	duration, _ := time.ParseDuration(offset)
	currentTime := time.Now().UTC().Add(duration)

	offset = strings.Replace(captures[offsetCapture], "+", "-", 1)
	fileName = "A" + captures[fileCreateDateCapture] + "." + captures[fileCreateTimeCapture] + offset
	fileName += "-" + fmt.Sprintf("%02d", currentTime.Hour()) + fmt.Sprintf("%02d", currentTime.Minute()) + offset
	fileName += "_" + foreignID
	if captures[seqNoCapture] != "" {
		fileName += "_" + captures[seqNoCapture]
	}
	log.Debug("renamed filename:", fileName)
	return fileName
}
