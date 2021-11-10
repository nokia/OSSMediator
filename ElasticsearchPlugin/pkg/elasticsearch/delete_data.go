/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"elasticsearchplugin/pkg/config"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//CleanUp deletes older data and indices from elasticsearch every day at 1 o'clock.
func CleanUp(esConf config.ElasticsearchConf) {
	cleanup(esConf)
	timer := time.NewTimer(getNextTickDuration())
	for {
		<-timer.C
		log.Info("Triggered data cleanup of elasticsearch")
		cleanup(esConf)
		timer.Reset(getNextTickDuration())
	}
}

func cleanup(esConf config.ElasticsearchConf) {
	log.Info("Triggered data cleanup of elasticsearch")
	deletionTime := currentTime().AddDate(0, 0, -1*esConf.DataRetentionDuration).UTC().Format(timestampFormat)
	deleteData(indexList, deletionTime, esConf)
	indicesToDelete := getOldIndices(esConf)
	deleteIndices(indicesToDelete, esConf)
}

func getNextTickDuration() time.Duration {
	currTime := currentTime()
	nextTick := time.Date(currTime.Year(), currTime.Month(), currTime.Day(), deletionHour, 0, 0, 0, time.Local)
	if nextTick.Before(currTime) {
		nextTick = nextTick.AddDate(0, 0, 1)
	}
	return nextTick.Sub(currentTime())
}

func deleteData(indices []string, deletionTime string, esConf config.ElasticsearchConf) {
	elkURL := esConf.URL + "/" + strings.Join(indices, ",") + elkDeleteAPI
	query := strings.Replace(deleteQuery, "TIMESTAMP", deletionTime, -1)
	queryParams := make(map[string]string)
	queryParams[elkWaitQueryParam] = "false"
	queryParams[elkIgnoreUnavailable] = "true"
	_, err := httpCall(http.MethodPost, elkURL, esConf.User, esConf.Password, query, queryParams)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "url": elkURL, "delete_from_time": deletionTime}).Error("Unable to delete data from elasticsearch")
	} else {
		log.WithFields(log.Fields{"url": elkURL, "delete_from_time": deletionTime}).Info("Data deleted from elasticsearch successfully.")
	}
}

func deleteIndices(indices []string, esConf config.ElasticsearchConf) {
	if len(indices) == 0 {
		return
	}
	log.Infof("Index patterns to delete %v", indices)
	elkURL := esConf.URL + "/" + strings.Join(indices, ",")
	queryParams := make(map[string]string)
	queryParams[elkIgnoreUnavailable] = "true"

	//closing the indices
	_, err := httpCall(http.MethodPost, elkURL+"/"+"_close", esConf.User, esConf.Password, "", queryParams)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "url": elkURL}).Error("unable to close indices")
		return
	}

	//deleting the indices
	_, err = httpCall(http.MethodDelete, elkURL, esConf.User, esConf.Password, "", queryParams)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "url": elkURL}).Error("unable to delete indices")
		return
	}
	log.Infof("Old indices deleted")
}

func getOldIndices(esConf config.ElasticsearchConf) []string {
	var indicesToDelete []string
	if esConf.DataRetentionDuration <= 31 {
		return indicesToDelete
	}

	indicesPattern := []string{"4g-pm*", "5g-pm*"}
	elkURL := esConf.URL + elkCatIndicesAPI + strings.Join(indicesPattern, ",") + "?h=index"
	//fetch the indices
	respBody, err := httpCall(http.MethodGet, elkURL, esConf.User, esConf.Password, "", nil)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "url": elkURL}).Error("Unable to fetch indices, exiting delete old indices process")
		return indicesToDelete
	}

	indices := strings.Split(string(respBody), "\n")
	delTime := time.Now().AddDate(0, 0, -1*esConf.DataRetentionDuration)
	delMonth := int(delTime.Month())
	delYear := delTime.Year()

	indicesPatternToDelete := make(map[string]struct{})
	for _, index := range indices {
		if index == "" {
			continue
		}
		splits := strings.Split(index, "-")
		indexCreationMonth, _ := strconv.Atoi(splits[len(splits)-2])
		indexCreationYear, _ := strconv.Atoi(splits[len(splits)-1])
		if indexCreationYear < delYear || (indexCreationYear == delYear && indexCreationMonth <= delMonth) {
			idPattern := "*" + strconv.Itoa(indexCreationMonth) + "-" + strconv.Itoa(indexCreationYear)
			indicesPatternToDelete[idPattern] = struct{}{}
		}
	}
	for index := range indicesPatternToDelete {
		indicesToDelete = append(indicesToDelete, index)
	}
	return indicesToDelete
}
