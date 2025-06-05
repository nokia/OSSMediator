/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"elasticsearchplugin/pkg/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SetConfig Set elasticsearch configuration for NDAC indices
func SetConfig(esConf config.ElasticsearchConf) {
	waitForElasticsearch(esConf)

	log.Info("Adding default OpenSearch settings")
	err := enableWildcardDeletion(esConf)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Unable to enable wildcard deletion on elasticsearch")
	}
	err = createDACIndexTemplate(esConf)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Unable to create index template for DAC indices")
	}
	err = setClusterSetting(esConf)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Unable to set cluster level settings")
	}
	log.Info("Default OpenSearch settings done")
}

// Check if Elasticsearch is reachable
func waitForElasticsearch(esConf config.ElasticsearchConf) {
	for {
		_, err := httpCall(http.MethodGet, esConf.URL, esConf.User, esConf.Password, nil, nil, defaultTimeout)
		if err == nil {
			log.Info("OpenSearch is reachable")
			return
		}
		log.WithFields(log.Fields{"Error": err}).Info("Waiting for Elasticsearch to be reachable...")
		time.Sleep(5 * time.Second)
	}
}

// enable wildcard deletion
func enableWildcardDeletion(esConf config.ElasticsearchConf) error {
	elkURL := esConf.URL + elkClusterSettingAPI
	query := `{"transient": {"action.destructive_requires_name": false}}`
	_, err := httpCall(http.MethodPut, elkURL, esConf.User, esConf.Password, &query, nil, defaultTimeout)
	return err
}

// Set number of shards used by dac indices to 1
func createDACIndexTemplate(esConf config.ElasticsearchConf) error {
	elkURL := esConf.URL + "/_index_template/dac-index"
	query := `{"index_patterns": ["4g-pm*", "5g-pm*", "ixr-pm", "edge-pm", "core-pm", "radio-fm", "dac-fm", "core-fm", "application-fm", "ixr-fm", "nhg-data", "sims-data", "account-sims-data", "ap-sims-data"],"template": {"settings": {"index.number_of_shards": 1}}}`
	_, err := httpCall(http.MethodPut, elkURL, esConf.User, esConf.Password, &query, nil, defaultTimeout)
	return err
}

// Set max_shards_per_node, search.max_buckets and script.max_compilations_rate cluster settings
func setClusterSetting(esConf config.ElasticsearchConf) error {
	elkURL := esConf.URL + elkClusterSettingAPI
	query := `{"persistent" : {"cluster.max_shards_per_node": "MAX_SHARDS", "search.max_buckets": "100000", "script.max_compilations_rate": "2000/5m"}}`
	query = strings.Replace(query, "MAX_SHARDS", strconv.Itoa(esConf.MaxShardsPerNode), -1)
	_, err := httpCall(http.MethodPut, elkURL, esConf.User, esConf.Password, &query, nil, defaultTimeout)
	return err
}
