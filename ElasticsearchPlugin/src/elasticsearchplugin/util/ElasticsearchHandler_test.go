/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	elasticsearchURL = "http://127.0.0.1:9200"
	testPMData       = `[
  {
    "_id": "EB184613063::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41532_NE-LNBTS-41532_LNCEL-2",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:09.953273425Z",
      "Dn": "NE-MRBTS-41532/NE-LNBTS-41532/LNCEL-2",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184613063::NOKIA::FW2QQD",
      "nename": "GTL Dallas Texas  eNB4430 DevKit#3 PICO",
      "neserialno": "EB184613063",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533_LNCEL-2",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:10.120112846Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533/LNCEL-2",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533_LNCEL-1",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:10.120254943Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533/LNCEL-1",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_PLMN-PLMN_MCC-999_MNC-40",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:10.120258344Z",
      "Dn": "PLMN-PLMN/MCC-999/MNC-40",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "MNC",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184613023::NOKIA::FW2QQD_2019-12-04T01:45:00.000-07:00:00_NE-MRBTS-46830_NE-LNBTS-46830_LNCEL-0",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:07.139687419Z",
      "Dn": "NE-MRBTS-46830/NE-LNBTS-46830/LNCEL-0",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-68-n-1-oam",
      "edge_id": "8K6F2R2::DELL::DEB-5000",
      "edgename": "ANTz",
      "ne_hw_id": "EB184613023::NOKIA::FW2QQD",
      "nename": "GTL SanFrancisco eNB46830 DevKit#1 PICO",
      "neserialno": "EB184613023",
      "nhg_id": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgid": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgname": "DevKit1",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T01:45:00.000-07:00:00"
    }
  },
  {
    "_id": "EB184613023::NOKIA::FW2QQD_2019-12-04T01:45:00.000-07:00:00_PLMN-PLMN_MCC-999_MNC-40",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:07.13969022Z",
      "Dn": "PLMN-PLMN/MCC-999/MNC-40",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "MNC",
      "edge_hostname": "24-68-n-1-oam",
      "edge_id": "8K6F2R2::DELL::DEB-5000",
      "edgename": "ANTz",
      "ne_hw_id": "EB184613023::NOKIA::FW2QQD",
      "nename": "GTL SanFrancisco eNB46830 DevKit#1 PICO",
      "neserialno": "EB184613023",
      "nhg_id": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgid": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgname": "DevKit1",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T01:45:00.000-07:00:00"
    }
  },
  {
    "_id": "EB193610864::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_NE-MRBTS-41535_NE-LNBTS-41535_LNCEL-1",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:09.260551536Z",
      "Dn": "NE-MRBTS-41535/NE-LNBTS-41535/LNCEL-1",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB193610864::NOKIA::FW2QQD",
      "nename": "EB193610864",
      "neserialno": "EB193610864",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB193610864::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_PLMN-PLMN_MCC-999_MNC-40",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:09.260554636Z",
      "Dn": "PLMN-PLMN/MCC-999/MNC-40",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "MNC",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB193610864::NOKIA::FW2QQD",
      "nename": "EB193610864",
      "neserialno": "EB193610864",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533_LNCEL-2",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:10.50628606Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533/LNCEL-2",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533_LNCEL-1",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:10.506385819Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533/LNCEL-1",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184613023::NOKIA::FW2QQD_2019-12-04T02:00:00.000-07:00:00_NE-MRBTS-46830_NE-LNBTS-46830_LNCEL-0",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:06.632393947Z",
      "Dn": "NE-MRBTS-46830/NE-LNBTS-46830/LNCEL-0",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "LNCEL",
      "edge_hostname": "24-68-n-1-oam",
      "edge_id": "8K6F2R2::DELL::DEB-5000",
      "edgename": "ANTz",
      "ne_hw_id": "EB184613023::NOKIA::FW2QQD",
      "nename": "GTL SanFrancisco eNB46830 DevKit#1 PICO",
      "neserialno": "EB184613023",
      "nhg_id": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgid": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgname": "DevKit1",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:00:00.000-07:00:00"
    }
  },
  {
    "_id": "EB184613023::NOKIA::FW2QQD_2019-12-04T02:00:00.000-07:00:00_PLMN-PLMN_MCC-999_MNC-40",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:06.632397933Z",
      "Dn": "PLMN-PLMN/MCC-999/MNC-40",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "MNC",
      "edge_hostname": "24-68-n-1-oam",
      "edge_id": "8K6F2R2::DELL::DEB-5000",
      "edgename": "ANTz",
      "ne_hw_id": "EB184613023::NOKIA::FW2QQD",
      "nename": "GTL SanFrancisco eNB46830 DevKit#1 PICO",
      "neserialno": "EB184613023",
      "nhg_id": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgid": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgname": "DevKit1",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:00:00.000-07:00:00"
    }
  },
  {
    "_id": "EB184613063::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_PLMN-PLMN_MCC-999_MNC-40",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:10.111151872Z",
      "Dn": "PLMN-PLMN/MCC-999/MNC-40",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "MNC",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184613063::NOKIA::FW2QQD",
      "nename": "GTL Dallas Texas  eNB4430 DevKit#3 PICO",
      "neserialno": "EB184613063",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212817::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_PLMN-PLMN_MCC-999_MNC-40",
    "_index": "edge-ne3s-pm-lte_x2_gnb_per_lncel-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:10.254992475Z",
      "Dn": "PLMN-PLMN/MCC-999/MNC-40",
      "LTE_X2_gNB_per_LNCEL_M8080C0": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C1": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C11": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C12": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C13": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C14": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C17": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C18": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C19": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C2": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C20": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C3": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C4": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C5": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C6": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C7": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C8": 0,
      "LTE_X2_gNB_per_LNCEL_M8080C9": 0,
      "ObjectType": "MNC",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212817::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41534 #5 PICO",
      "neserialno": "EB184212817",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB193610864::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41535_NE-LNBTS-41535_LNADJ-254",
    "_index": "edge-ne3s-pm-lte_x2_sctp_statistics-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:08.297881304Z",
      "Dn": "NE-MRBTS-41535/NE-LNBTS-41535/LNADJ-254",
      "LTE_X2_SCTP_Statistics_M8036C0": 0,
      "LTE_X2_SCTP_Statistics_M8036C1": 0,
      "LTE_X2_SCTP_Statistics_M8036C2": 0,
      "LTE_X2_SCTP_Statistics_M8036C3": 0,
      "LTE_X2_SCTP_Statistics_M8036C4": 0,
      "LTE_X2_SCTP_Statistics_M8036C5": 0,
      "LTE_X2_SCTP_Statistics_M8036C6": 0,
      "ObjectType": "LNADJ",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB193610864::NOKIA::FW2QQD",
      "nename": "EB193610864",
      "neserialno": "EB193610864",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB193610864::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41535_NE-LNBTS-41535_LNADJ-253",
    "_index": "edge-ne3s-pm-lte_x2_sctp_statistics-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:08.297883342Z",
      "Dn": "NE-MRBTS-41535/NE-LNBTS-41535/LNADJ-253",
      "LTE_X2_SCTP_Statistics_M8036C0": 0,
      "LTE_X2_SCTP_Statistics_M8036C1": 0,
      "LTE_X2_SCTP_Statistics_M8036C2": 0,
      "LTE_X2_SCTP_Statistics_M8036C3": 0,
      "LTE_X2_SCTP_Statistics_M8036C4": 0,
      "LTE_X2_SCTP_Statistics_M8036C5": 0,
      "LTE_X2_SCTP_Statistics_M8036C6": 0,
      "ObjectType": "LNADJ",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB193610864::NOKIA::FW2QQD",
      "nename": "EB193610864",
      "neserialno": "EB193610864",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184613063::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41532_NE-LNBTS-41532_LNADJ-254",
    "_index": "edge-ne3s-pm-lte_x2_sctp_statistics-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:09.945765382Z",
      "Dn": "NE-MRBTS-41532/NE-LNBTS-41532/LNADJ-254",
      "LTE_X2_SCTP_Statistics_M8036C0": 0,
      "LTE_X2_SCTP_Statistics_M8036C1": 0,
      "LTE_X2_SCTP_Statistics_M8036C2": 0,
      "LTE_X2_SCTP_Statistics_M8036C3": 0,
      "LTE_X2_SCTP_Statistics_M8036C4": 0,
      "LTE_X2_SCTP_Statistics_M8036C5": 0,
      "LTE_X2_SCTP_Statistics_M8036C6": 0,
      "ObjectType": "LNADJ",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184613063::NOKIA::FW2QQD",
      "nename": "GTL Dallas Texas  eNB4430 DevKit#3 PICO",
      "neserialno": "EB184613063",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533_LNADJ-254",
    "_index": "edge-ne3s-pm-lte_x2_sctp_statistics-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:10.115319679Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533/LNADJ-254",
      "LTE_X2_SCTP_Statistics_M8036C0": 0,
      "LTE_X2_SCTP_Statistics_M8036C1": 0,
      "LTE_X2_SCTP_Statistics_M8036C2": 0,
      "LTE_X2_SCTP_Statistics_M8036C3": 0,
      "LTE_X2_SCTP_Statistics_M8036C4": 0,
      "LTE_X2_SCTP_Statistics_M8036C5": 0,
      "LTE_X2_SCTP_Statistics_M8036C6": 0,
      "ObjectType": "LNADJ",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212817::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_NE-MRBTS-41534_NE-LNBTS-41534_LNADJ-254",
    "_index": "edge-ne3s-pm-lte_x2_sctp_statistics-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:10.246668123Z",
      "Dn": "NE-MRBTS-41534/NE-LNBTS-41534/LNADJ-254",
      "LTE_X2_SCTP_Statistics_M8036C0": 0,
      "LTE_X2_SCTP_Statistics_M8036C1": 0,
      "LTE_X2_SCTP_Statistics_M8036C2": 0,
      "LTE_X2_SCTP_Statistics_M8036C3": 0,
      "LTE_X2_SCTP_Statistics_M8036C4": 0,
      "LTE_X2_SCTP_Statistics_M8036C5": 0,
      "LTE_X2_SCTP_Statistics_M8036C6": 0,
      "ObjectType": "LNADJ",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212817::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41534 #5 PICO",
      "neserialno": "EB184212817",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184613063::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41532_NE-LNBTS-41532",
    "_index": "edge-ne3s-pm-lte_x2ap-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:09.900561757Z",
      "Dn": "NE-MRBTS-41532/NE-LNBTS-41532",
      "LTE_X2AP_M8022C0": 0,
      "LTE_X2AP_M8022C1": 0,
      "ObjectType": "LNBTS",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184613063::NOKIA::FW2QQD",
      "nename": "GTL Dallas Texas  eNB4430 DevKit#3 PICO",
      "neserialno": "EB184613063",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533",
    "_index": "edge-ne3s-pm-lte_x2ap-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:10.11036389Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533",
      "LTE_X2AP_M8022C0": 0,
      "LTE_X2AP_M8022C1": 0,
      "ObjectType": "LNBTS",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184613023::NOKIA::FW2QQD_2019-12-04T01:45:00.000-07:00:00_NE-MRBTS-46830_NE-LNBTS-46830",
    "_index": "edge-ne3s-pm-lte_x2ap-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:07.134160113Z",
      "Dn": "NE-MRBTS-46830/NE-LNBTS-46830",
      "LTE_X2AP_M8022C0": 0,
      "LTE_X2AP_M8022C1": 0,
      "ObjectType": "LNBTS",
      "edge_hostname": "24-68-n-1-oam",
      "edge_id": "8K6F2R2::DELL::DEB-5000",
      "edgename": "ANTz",
      "ne_hw_id": "EB184613023::NOKIA::FW2QQD",
      "nename": "GTL SanFrancisco eNB46830 DevKit#1 PICO",
      "neserialno": "EB184613023",
      "nhg_id": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgid": "efba1743-6873-4e2e-58f1-9fbf7a66cd64",
      "nhgname": "DevKit1",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T01:45:00.000-07:00:00"
    }
  },
  {
    "_id": "EB193610864::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_NE-MRBTS-41535_NE-LNBTS-41535",
    "_index": "edge-ne3s-pm-lte_x2ap-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:09.252627749Z",
      "Dn": "NE-MRBTS-41535/NE-LNBTS-41535",
      "LTE_X2AP_M8022C0": 0,
      "LTE_X2AP_M8022C1": 0,
      "ObjectType": "LNBTS",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB193610864::NOKIA::FW2QQD",
      "nename": "EB193610864",
      "neserialno": "EB193610864",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533",
    "_index": "edge-ne3s-pm-lte_x2ap-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:10.497173615Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533",
      "LTE_X2AP_M8022C0": 0,
      "LTE_X2AP_M8022C1": 0,
      "ObjectType": "LNBTS",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T02:45:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533_FTM-1_IPNO-1_INTP-1",
    "_index": "edge-ne3s-pm-ntp_stats-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:00:10.121187201Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533/FTM-1/IPNO-1/INTP-1",
      "NTP_stats_M51151C0": 0,
      "NTP_stats_M51151C1": 0,
      "NTP_stats_M51151C2": 0,
      "ObjectType": "INTP",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T02:45:00.000-06:00:00"
    }
  },
  {
    "_id": "EB184212819::NOKIA::FW2QQD_2019-12-04T03:00:00.000-06:00:00_NE-MRBTS-41533_NE-LNBTS-41533_FTM-1_IPNO-1_INTP-1",
    "_index": "edge-ne3s-pm-ntp_stats-12-2019",
    "_source": {
      "@timestamp": "2019-12-04T09:15:10.507215752Z",
      "Dn": "NE-MRBTS-41533/NE-LNBTS-41533/FTM-1/IPNO-1/INTP-1",
      "NTP_stats_M51151C0": 0,
      "NTP_stats_M51151C1": 0,
      "NTP_stats_M51151C2": 0,
      "ObjectType": "INTP",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212819::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41533 #4 PICO",
      "neserialno": "EB184212819",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss",
      "timestamp": "2019-12-04T03:00:00.000-06:00:00"
    }
  }
]
`

	testFMData = `[
  {
    "_id": "EB184212817::NOKIA::FW2QQD_MRBTS-41534/LNBTS-41534/LNCEL-1_16_7655_2019-12-04 03:52:27 -0600 -0600_{    }",
    "_index": "edge-fm",
    "_source": {
      "@timestamp": "2019-12-04T09:52:29.671205438Z",
      "AdditionalText": "SasServerCommunicationFailureAl6832;serial_no.=EB184212817;supplAlarmInfo:Response code: 700 Unable to connect to SAS Server and alternate server available;0",
      "AlarmIdentifier": "16",
      "AlarmState": 1,
      "AlarmText": "SAS Server Communication Failure",
      "Dn": "MRBTS-41534/LNBTS-41534/LNCEL-1",
      "EventTime": "2019-12-04T03:52:27-06:00",
      "EventType": "qualityOfService",
      "LastUpdatedTime": "2019-12-04T03:52:27-06:00",
      "ManagedObjectInstance": "MRBTS-41534/LNBTS-41534/LNCEL-1",
      "NotificationType": "alarmNew",
      "ProbableCause": "",
      "Severity": "minor",
      "SpecificProblem": "7655",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212817::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41534 #5 PICO",
      "neserialno": "EB184212817",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss"
    }
  },
  {
    "_id": "EB184212817::NOKIA::FW2QQD_MRBTS-41534/LNBTS-41534/LNCEL-1_17_7655_2019-12-04 03:52:27 -0600 -0600_{    }",
    "_index": "edge-fm",
    "_source": {
      "@timestamp": "2019-12-04T09:52:29.671225366Z",
      "AdditionalText": "SasHeartbeatFailureAl6829;serial_no.=EB184212817;supplAlarmInfo:Unable to connect to SAS Server and alternate server available;0",
      "AlarmIdentifier": "17",
      "AlarmState": 1,
      "AlarmText": "SAS Heartbeat Failure",
      "Dn": "MRBTS-41534/LNBTS-41534/LNCEL-1",
      "EventTime": "2019-12-04T03:52:27-06:00",
      "EventType": "qualityOfService",
      "LastUpdatedTime": "2019-12-04T03:52:27-06:00",
      "ManagedObjectInstance": "MRBTS-41534/LNBTS-41534/LNCEL-1",
      "NotificationType": "alarmNew",
      "ProbableCause": "",
      "Severity": "minor",
      "SpecificProblem": "7655",
      "edge_hostname": "24-15-n-1-oam",
      "edge_id": "102H2R2::DELL::DEB-5000",
      "edgename": "Burnet Edge",
      "ne_hw_id": "EB184212817::NOKIA::FW2QQD",
      "nename": "GTL Burnet eNB41534 #5 PICO",
      "neserialno": "EB184212817",
      "nhg_id": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgid": "84713035-ac4c-482a-5e64-f8d2db5029aa",
      "nhgname": "GTL Burnet",
      "tag": "moss",
      "tag_value": "moss"
    }
  }
]
`
)

func TestPushPMDataToElasticsearch(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	fileName := "./pm_data.json"
	err := createTestData(fileName, testPMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./pm_data.json", elasticsearchURL)
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}

func TestPushFMDataToElasticsearch(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	fileName := "./fm_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./fm_data.json", elasticsearchURL)
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}

func TestPushFMDataToInvalidElasticsearch(t *testing.T) {
	elasticsearchURL = "http://localhost:12345"
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
		elasticsearchURL = "http://127.0.0.1:9200"
	}()
	fileName := "./fm_data.json"
	err := createTestData(fileName, testFMData)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fileName)

	pushDataToElasticsearch("./fm_data.json", elasticsearchURL)
	if !strings.Contains(buf.String(), "Unable to push data to elasticsearch, will be retried later") {
		t.Fail()
	}
}

func TestPushFailedDataToElasticsearch(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	fileName := "./fm_data.json"
	retryTicker = time.NewTicker(2 * time.Millisecond)
	PushFailedDataToElasticsearch(elasticsearchURL)
	time.Sleep(1 * time.Second)
	t.Log(buf.String())
	if !strings.Contains(buf.String(), "Data from "+fileName+" pushed to elasticsearch successfully") {
		t.Fail()
	}
}
