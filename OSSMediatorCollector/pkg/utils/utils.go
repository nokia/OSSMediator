/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package utils

import (
	"collector/pkg/config"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"
)

var (
	errorPasswordRead         = errors.New("unable to read password file")
	errorPasswordDecoding     = errors.New("unable to decode password")
	errorPasswordFileNotFound = errors.New("secret file not found")
)

// truncates seconds from time
func truncateSeconds(t time.Time) time.Time {
	return t.Truncate(60 * time.Second)
}

// GetTimeInterval returns the start_time and end_time for API calls.
// It reads start_time from file where last received data's event time is stored
func GetTimeInterval(user *config.User, api *config.APIConf, nhgID string) (string, string) {
	currentTime := CurrentTime().Add(time.Duration(-10) * time.Second)
	//calculating 15 minutes time frame
	diff := currentTime.Minute() - (currentTime.Minute() / api.Interval * api.Interval) + api.Interval
	begTime := currentTime.Add(time.Duration(-1*diff) * time.Minute)
	endTime := begTime.Add(time.Duration(api.Interval) * time.Minute)
	startTime := getLastReceivedDataTime(user, api, nhgID)
	if startTime == "" {
		startTime = truncateSeconds(begTime).Format(time.RFC3339)
	}
	if api.SyncDuration > 0 {
		stTime, _ := time.Parse(time.RFC3339, startTime)
		gap := endTime.Sub(stTime)
		if gap < (time.Duration(api.SyncDuration) * time.Minute) {
			stTime = endTime.Add(time.Duration(-1*api.SyncDuration) * time.Minute)
			startTime = truncateSeconds(stTime).Format(time.RFC3339)
		}
	}
	return startTime, truncateSeconds(endTime).Format(time.RFC3339)
}

func ReadPassword(email string) (string, error) {
	passwordFile := ".secret/." + email
	var bytePassword []byte
	if fileExists(passwordFile) {
		data, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", err
		}
		if len(data) == 0 {
			return "", errorPasswordRead
		}
		bytePassword, err = base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return "", errorPasswordDecoding
		}
	} else {
		return "", errorPasswordFileNotFound
	}
	return string(bytePassword), nil
}

func ReadSessionToken(email string) (string, error) {
	passwordFile := ".secret/." + email
	var bytePassword []byte
	var byteAccessToken []byte
	var byteRefreshToken []byte
	if fileExists(passwordFile) {
		data, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", err
		}
		if len(data) == 0 {
			return "", errorPasswordRead
		}
		sessionToken := strings.Split(string(data), "\n")
		byteAccessToken, err = base64.StdEncoding.DecodeString(sessionToken[0])
		if err != nil {
			return "", errorPasswordDecoding
		}
		byteRefreshToken, err = base64.StdEncoding.DecodeString(sessionToken[1])
		if err != nil {
			return "", errorPasswordDecoding
		}
		bytePassword = []byte(string(byteAccessToken) + "\n" + string(byteRefreshToken))
	} else {
		return "", errorPasswordFileNotFound
	}
	return string(bytePassword), nil
}
