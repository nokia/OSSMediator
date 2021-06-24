/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package utils

import (
	"collector/config"
	"time"
)

//truncates seconds from time
func truncateSeconds(t time.Time) time.Time {
	return t.Truncate(60 * time.Second)
}

//GetTimeInterval returns the start_time and end_time for API calls.
//It reads start_time from file where last received data's event time is stored
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
	endTime = endTime.Add(-1 * time.Minute)
	return startTime, endTime.Format(time.RFC3339)
}
