/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"elasticsearchplugin/pkg/config"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"path"
	"strings"
	"time"
)

//to read sims api response
type simsResponse struct {
	Subsc []struct {
		ApnID                 string                  `json:"apn_id"`
		ApnName               string                  `json:"apn_name"`
		IccID                 string                  `json:"icc_id"`
		Imei                  string                  `json:"imei"`
		Imsi                  string                  `json:"imsi"`
		ImsiAlias             string                  `json:"imsi_alias"`
		LastKnownAccessPoint  interface{}             `json:"last_known_access_point"`
		PrivateNetworkDetails []PrivateNetworkDetails `json:"private_network_details"`
		Qos                   string                  `json:"qos"`
		QosGroupID            string                  `json:"qos_group_id"`
		QosProfileName        string                  `json:"qos_profile_name"`
		StaticIP              string                  `json:"static_ip"`
	} `json:"subsc"`
}

//sims details for elasticsearch data
type simsDetails struct {
	ApnID                 string                `json:"apn_id"`
	ApnName               string                `json:"apn_name"`
	IccID                 string                `json:"icc_id"`
	Imei                  string                `json:"imei"`
	Imsi                  string                `json:"imsi"`
	ImsiAlias             string                `json:"imsi_alias"`
	LastKnownAccessPoint  interface{}           `json:"last_known_access_point"`
	PrivateNetworkDetails PrivateNetworkDetails `json:"private_network_details"`
	Qos                   string                `json:"qos"`
	QosGroupID            string                `json:"qos_group_id"`
	QosProfileName        string                `json:"qos_profile_name"`
	StaticIP              string                `json:"static_ip"`
	Timestamp             time.Time             `json:"timestamp"`
}

type PrivateNetworkDetails struct {
	FeatureStatus     interface{} `json:"feature_status"`
	NhgAlias          string      `json:"nhg_alias"`
	NhgID             string      `json:"nhg_id"`
	Status            string      `json:"status"`
	StatusDescription string      `json:"status_description"`
	UeIP              string      `json:"ue_ip"`
	UeReportTimestamp time.Time   `json:"ue_report_timestamp"`
	UeStatus          string      `json:"ue_status"`
}

//to read ap-sims api response
type apSimsResponse []struct {
	AccessPointHwID string        `json:"access_point_hw_id"`
	ImsiDetails     []ImsiDetails `json:"imsi_details"`
}

type ImsiDetails struct {
	IccID             string    `json:"icc_id"`
	Imsi              string    `json:"imsi"`
	ImsiAlias         string    `json:"imsi_alias"`
	UeReportTimestamp time.Time `json:"ue_report_timestamp"`
}

//ap-sims details for elasticsearch data
type apSimsDetails struct {
	AccessPointHwID string      `json:"access_point_hw_id"`
	ImsiDetails     ImsiDetails `json:"imsi_details"`
	Timestamp       time.Time   `json:"timestamp"`
}

func pushAPSimsData(filePath string, esConf config.ElasticsearchConf) {
	//deleting data from ap-sims-data index
	deletionTime := currentTime().Add(-15 * time.Minute).UTC().Format(timestampFormat)
	deleteData([]string{indexMetaData[apSimsData]}, deletionTime, esConf)

	elkURL := esConf.URL + elkBulkAPI
	log.Infof("Pushing data from %s to elasticsearch", filePath)
	data, err := readFile(filePath)
	if err != nil {
		return
	}

	fileName := path.Base(filePath)
	keys := strings.Split(fileName, "_")
	metric := strings.ToLower(keys[0])
	accID := strings.ToLower(keys[1])

	var apSims apSimsResponse
	err = json.Unmarshal(data, &apSims)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}

	var postData string
	index := indexMetaData[metric]
	currTime := time.Now().UTC()
	for _, apInfo := range apSims {
		data := apSimsDetails{
			AccessPointHwID: apInfo.AccessPointHwID,
			Timestamp:       currTime,
		}
		for _, imsiInfo := range apInfo.ImsiDetails {
			data.ImsiDetails = imsiInfo
			id := strings.Join([]string{metric, accID, data.AccessPointHwID, data.ImsiDetails.Imsi}, "_")
			source, _ := json.Marshal(data)
			postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
			postData += string(source) + "\n"
		}
	}
	pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
}

func pushSimsData(filePath string, esConf config.ElasticsearchConf) {
	//deleting data from sims-data and account-sims-data indices
	deletionTime := currentTime().Add(-15 * time.Minute).UTC().Format(timestampFormat)
	deleteData([]string{indexMetaData[simsData], indexMetaData[accountSimsData]}, deletionTime, esConf)

	elkURL := esConf.URL + elkBulkAPI
	log.Infof("Pushing data from %s to elasticsearch", filePath)
	data, err := readFile(filePath)
	if err != nil {
		return
	}

	fileName := path.Base(filePath)
	keys := strings.Split(fileName, "_")
	metric := strings.ToLower(keys[0])
	accID := strings.ToLower(keys[1])

	var sims simsResponse
	err = json.Unmarshal(data, &sims)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}

	var postData string
	index := indexMetaData[metric]
	currTime := time.Now().UTC()
	for _, subsc := range sims.Subsc {
		data := simsDetails{
			ApnID:                subsc.ApnID,
			ApnName:              subsc.ApnName,
			IccID:                subsc.IccID,
			Imei:                 subsc.Imei,
			Imsi:                 subsc.Imsi,
			ImsiAlias:            subsc.ImsiAlias,
			LastKnownAccessPoint: subsc.LastKnownAccessPoint,
			Qos:                  subsc.Qos,
			QosGroupID:           subsc.QosGroupID,
			QosProfileName:       subsc.QosProfileName,
			StaticIP:             subsc.StaticIP,
			Timestamp:            currTime,
		}
		for _, networkDetails := range subsc.PrivateNetworkDetails {
			if networkDetails.NhgID == "" {
				continue
			}
			data.PrivateNetworkDetails = networkDetails
			id := strings.Join([]string{metric, accID, data.Imsi, data.PrivateNetworkDetails.NhgID}, "_")
			source, _ := json.Marshal(data)
			postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
			postData += string(source) + "\n"
		}
	}
	pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
}
