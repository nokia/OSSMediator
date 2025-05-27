/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package elasticsearch

import (
	"elasticsearchplugin/pkg/config"
	"net/http"
	"strings"
	"testing"
)

func TestSetConfigSuccess(t *testing.T) {
	esConf := config.ElasticsearchConf{
		URL:              elasticsearchURL,
		MaxShardsPerNode: 1000,
	}

	SetConfig(esConf)

	searchURL := elasticsearchURL + "/_cluster/settings"
	resp, err := httpCall(http.MethodGet, searchURL, "", "", nil, nil, defaultTimeout)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(string(resp), "\"destructive_requires_name\":\"false\"") {
		t.Fail()
	}
	if !strings.Contains(string(resp), "\"max_shards_per_node\":\"1000\"") {
		t.Fail()
	}
	if !strings.Contains(string(resp), "\"max_buckets\":\"100000\"") {
		t.Fail()
	}
	if !strings.Contains(string(resp), "\"max_compilations_rate\":\"2000/5m\"") {
		t.Fail()
	}
}

func TestEnableWildcardDeletionInvalidURL(t *testing.T) {
	esConf := config.ElasticsearchConf{
		URL: "http://invalid-url",
	}

	err := enableWildcardDeletion(esConf)
	if err == nil {
		t.Fail()
	}
}

func TestCreateDACIndexTemplateInvalidURL(t *testing.T) {
	esConf := config.ElasticsearchConf{
		URL: "http://invalid-url",
	}

	err := createDACIndexTemplate(esConf)
	if err == nil {
		t.Fail()
	}
}

func TestSetClusterSettingInvalidURL(t *testing.T) {
	esConf := config.ElasticsearchConf{
		URL:              "http://invalid-url",
		MaxShardsPerNode: 1000,
	}

	err := setClusterSetting(esConf)
	if err == nil {
		t.Fail()
	}
}

func TestCheck(t *testing.T) {
	esConf := config.ElasticsearchConf{
		URL:      "https://api.moss.m-dev.dev.srv.da.nsn-rdnet.net",
		User:     "mossos",
		Password: "mossos123",
	}
	waitForElasticsearch(esConf)
}
