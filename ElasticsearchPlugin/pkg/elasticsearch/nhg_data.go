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

type nhgResponse []struct {
	NhgID                 string `json:"nhg_id"`
	NhgAlias              string `json:"nhg_alias"`
	DeploymentType        string `json:"deployment_type"`
	DeployReadinessStatus int    `json:"deploy_readiness_status"`
	Clusters              []struct {
		ClusterID                    string    `json:"cluster_id"`
		ClusterAlias                 string    `json:"cluster_alias"`
		HwSet                        []HwSet   `json:"hw_set"`
		ClusterStatus                string    `json:"cluster_status"`
		Region                       string    `json:"region"`
		Location                     string    `json:"location"`
		Latitude                     float32   `json:"latitude"`
		Longitude                    float32   `json:"longitude"`
		NwConfigStatus               string    `json:"nw_config_status"`
		EdgeConnectionStatus         string    `json:"edge_connection_status"`
		ClusterSliceStatusModifyTime time.Time `json:"cluster_slice_status_modify_time"`
		EdgeIPAddress                string    `json:"edge_ip_address"`
		SrxIPAddress                 string    `json:"srx_ip_address"`
		SliceID                      string    `json:"slice_id"`
	} `json:"clusters"`
	NhgConfigStatus          string    `json:"nhg_config_status"`
	NhgSliceStatusModifyTime time.Time `json:"nhg_slice_status_modify_time"`
	NetworkType              string    `json:"network_type"`
	UsbInstallStatus         string    `json:"usb_install_status"`
	SLA                      string    `json:"sla"`
}

type nhgDetails struct {
	NhgID                    string    `json:"nhg_id"`
	NhgAlias                 string    `json:"nhg_alias"`
	DeploymentType           string    `json:"deployment_type"`
	DeployReadinessStatus    int       `json:"deploy_readiness_status"`
	Cluster                  Cluster   `json:"cluster"`
	NhgConfigStatus          string    `json:"nhg_config_status"`
	NhgSliceStatusModifyTime time.Time `json:"nhg_slice_status_modify_time"`
	NetworkType              string    `json:"network_type"`
	UsbInstallStatus         string    `json:"usb_install_status"`
	SLA                      string    `json:"sla"`
	Timestamp                time.Time `json:"timestamp"`
}

type Cluster struct {
	ClusterID                    string    `json:"cluster_id"`
	ClusterAlias                 string    `json:"cluster_alias"`
	HwSet                        HwSet     `json:"hw_set"`
	ClusterStatus                string    `json:"cluster_status"`
	Region                       string    `json:"region"`
	Location                     string    `json:"location"`
	Latitude                     float32   `json:"latitude"`
	Longitude                    float32   `json:"longitude"`
	NwConfigStatus               string    `json:"nw_config_status"`
	EdgeConnectionStatus         string    `json:"edge_connection_status"`
	ClusterSliceStatusModifyTime time.Time `json:"cluster_slice_status_modify_time"`
	EdgeIPAddress                string    `json:"edge_ip_address"`
	SrxIPAddress                 string    `json:"srx_ip_address"`
	SliceID                      string    `json:"slice_id"`
}

type HwSet struct {
	HwID     string `json:"hw_id"`
	HwType   string `json:"hw_type"`
	SerialNo string `json:"serial_no"`
}

func pushNHGData(filePath string, esConf config.ElasticsearchConf) {
	//deleting data from nhg-data index
	deletionTime := currentTime().Add(-15 * time.Minute).UTC().Format(timestampFormat)
	deleteData([]string{indexMetaData[nhgData]}, deletionTime, esConf)

	elkURL := esConf.URL + elkBulkAPI
	log.Infof("Pushing data from %s to elasticsearch", filePath)
	data, err := readFile(filePath)
	if err != nil {
		return
	}

	fileName := path.Base(filePath)
	keys := strings.Split(fileName, "_")
	metric := strings.ToLower(keys[0])
	user := strings.ToLower(keys[1])

	var nhgs nhgResponse
	err = json.Unmarshal(data, &nhgs)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Unable to unmarshal json data %s", filePath)
		return
	}
	var postData string
	index := indexMetaData[nhgData]
	currTime := time.Now().UTC()
	for _, nhg := range nhgs {
		data := nhgDetails{
			NhgID:                    nhg.NhgID,
			NhgAlias:                 nhg.NhgAlias,
			DeploymentType:           nhg.DeploymentType,
			DeployReadinessStatus:    nhg.DeployReadinessStatus,
			NhgConfigStatus:          nhg.NhgConfigStatus,
			NhgSliceStatusModifyTime: nhg.NhgSliceStatusModifyTime,
			NetworkType:              nhg.NetworkType,
			UsbInstallStatus:         nhg.UsbInstallStatus,
			SLA:                      nhg.SLA,
			Timestamp:                currTime,
		}
		for _, cluster := range nhg.Clusters {
			data.Cluster = Cluster{
				ClusterID:                    cluster.ClusterID,
				ClusterAlias:                 cluster.ClusterAlias,
				ClusterStatus:                cluster.ClusterStatus,
				Region:                       cluster.Region,
				Location:                     cluster.Location,
				Latitude:                     cluster.Latitude,
				Longitude:                    cluster.Longitude,
				NwConfigStatus:               cluster.NwConfigStatus,
				EdgeConnectionStatus:         cluster.EdgeConnectionStatus,
				ClusterSliceStatusModifyTime: cluster.ClusterSliceStatusModifyTime,
				EdgeIPAddress:                cluster.EdgeIPAddress,
				SrxIPAddress:                 cluster.SrxIPAddress,
				SliceID:                      cluster.SliceID,
			}
			for _, hwSet := range cluster.HwSet {
				data.Cluster.HwSet = hwSet
				id := strings.Join([]string{metric, user, data.NhgID, data.Cluster.ClusterID, data.Cluster.HwSet.HwID}, "_")
				source, _ := json.Marshal(data)
				postData += `{"index": {"_index": "` + index + `", "_id": "` + id + `"}}` + "\n"
				postData += string(source) + "\n"
			}
		}
	}
	pushData(elkURL, esConf.User, esConf.Password, postData, filePath)
}
