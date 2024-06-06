# Changelog
All notable changes to this project will be documented in this file.

# 4.3

NEW FEATURES:

* OSSMediator:
  * Added `-v` option to display the version of OSSMediator.
  * Updated `golang` version to `1.22.3`. 
* MediatorSetup:
  * Added `NDAC Radio 4G PM KPI MOCN Dashboard`.
  * Added QOS KPIs in `CORE 5G PM Dashboard` and `Radio 5G Dashboards`.

BUG FIXES:

* OSSMediatorCollector: Fix for data collection for ABAC user.

SECURITY UPDATE:

* OSSMediator:
  * Updating base image to `alpine:3.20.0`.
* MediatorSetup:
  * Updating OpenSearch version to `2.14.0`.
  * Updating Grafana version to `11.0.0`.

# 4.2

NEW FEATURES:

* OSSMediatorCollector:
  * Adding support for `MXIE` networks.
    * Added `/generic-network-groups` API to get the list of MXIE networks.
  * Add support to enable alarm notification for all alarms. User can add * wildcard in alarm filter config to enable all alarm notification. 
* MediatorSetup:
  * `All Alarm Dashboard` to view all alarms in single dashboard.
  * Added `UE throughput - Per IMSI` panel in `CORE PM` dashboards. 

REMOVED:

* MediatorSetup: Removed RADIO, CORE, DAC, IXR and APPLICATION specific alarm dashboards..

SECURITY UPDATE:

* OSSMediator:
  * Updating base image to `alpine:3.19.1`.
* MediatorSetup:
  * Updating OpenSearch version to `2.13.0`.
  * Updating Grafana version to `9.5.18`.

# 4.1

NEW FEATURES:

* OSSMediatorCollector:
  * Adding `aggregation` parameter for IXR,EDGE and CORE PM API.

IMPROVEMENTS:

* MediatorSetup:
  * Adding `Aggregation` parameter for IXR,EDGE and CORE PM Dashboard.
  * IXR and CORE PM Dashboard improvements.
  * Updated default Grafana version to `9.5.15`.

BUG FIXES:

* OSSMediatorCollector: Removing start_time and end_time from fmdata API for ACTIVE alarm_type to fetch all active alarms.

SECURITY UPDATE:

* OSSMediator:
  * Updating base image to `alpine:3.19.0`.

# 4.0

NEW FEATURES:

* OSSMediatorCollector:
  * Added support for `ABAC` users.
  * Added configuration `auth_type` to identify authorization type for user, `PASSWORD` for `RBAC` and `ADTOKEN` for `ABAC` user.
  * Added configuration `pretty_response` to enable/disable pretty json response file.

IMPROVEMENTS:

* MediatorSetup:
  * Dashboard improvements.
  * Updated OpenSearch version to `2.9.0`.

BUG FIXES:

* ElasticsearchPlugin: Added deletion of old data from core-pm, edge-pm and ixr-pm index.

SECURITY UPDATE:

* OSSMediator:
  * Updating base image to `alpine:3.18.3`.

# 3.8

NEW FEATURES:

* OSSMediator:
  * Added support for `IXR` pm and fm APIs.
* MediatorSetup:
  * Added dashboards to view `NDAC IXR PM KPI`, `NDAC IXR Alarms Active Dashboard` and `NDAC IXR Alarms History Dashboard` metrics.

IMPROVEMENTS:

* MediatorSetup:
  * Added `NHG Name` filter for `CORE PM Dashboards` and `EDGE PM Dashboard`.
  * Updated OpenSearch version to `2.8.0`.

# 3.7

NEW FEATURES:

* OSSMediator: 
  * Changes to support running OSSMediator in docker.
* OSSMediatorCollector:
  * Added `max_concurrent_process` parameter in config for concurrent call of PM/FM APIs.
* ElasticsearchPlugin:
  * Added `max_concurrent_process` parameter in config for concurrent push of PM/FM response to OpenSearch/Elasticsearch.
* MediatorSetup:
  * Added dashboards to view 5G CORE PM dashboard.
  * Added `mediator_docker_startup.sh` script to start OSSMediator setup with docker.
  * Added `data_migration.sh` script to migrate data from `elasticsearch` to `OpenSearch`.
  * Using `OpenSearch` for storing KPI data instead of `elasticsearch`.
  * Updated required grafana version to `9.2.6`.

IMPROVEMENTS:

* MediatorSetup:
  * Updated 4G CORE PM dashboard.

BUG FIXES:

* ElasticsearchPlugin: core-pm index has incorrect default mapping of long type instead of float.

REMOVED:

* MediatorSetup: Removed installation of elasticsearch for storing KPI data.

# 3.6

BUG FIXES:

* ElasticsearchPlugin: Fix for ignoring pm indices without month and year suffix.

# 3.5

IMPROVEMENTS:

* OSSMediatorCollector: Updated `storesecret` module.
* MediatorSetup:
  * Added `grafana_cleanup.sh` script to remove NDAC dashboards from grafana.
  * Update to create the grafana dashboards inside `Nokia-DAC` directory instead of `General` directory.

BUG FIXES:

* ElasticsearchPlugin: Fix to clear `RADIO` alarms.
* OSSMediatorCollector: Fix for data loss when APIs were called by multiple users.

# 3.4

NEW FEATURES:

* OSSMediatorCollector:
  * Added support for getting `EDGE` and `CORE` pm data.
  * Added support for getting `APPLICATION` alarms.

* ElasticsearchPlugin:
  * Changes for pushing `EDGE`, `CORE` pm data and `APPLICATION` alarms to elasticsearch.

* MediatorSetup:
  * Added dashboards to view ACTIVE/HISTORY `APPLICATION` alarms.
  * Added dashboards to view `EDGE` and `CORE` pm KPIs.
  * Added `storesecret.sh` script to take user's password as input and store secretly for API access.
  * Added `startup_without_elasticsearch.sh` script to start OSSMediator setup with pre-installed elasticsearch.
  * Updated required grafana version to `8.5.9`.

IMPROVEMENTS:

* OSSMediatorCollector: Reading user's password from secret file.
* ElasticsearchPlugin: Document updates.
* MediatorSetup: Document updates for installation with pre-installed elasticsearch.

REMOVED:

* OSSMediatorCollector: Removed reading base64 encoded password from config file.

# 3.3

NEW FEATURES:

* Added PM multifire eNB dashboards.

BUG FIXES:

* ElasticsearchPlugin: Memory leak bug fix.

# 3.2

NEW FEATURES:

* OSSMediatorCollector:
  * Added support for access point SIM API.
* ElasticsearchPlugin:
  * Changes for pushing access point SIM data to elasticsearch.

IMPROVEMENTS:

* MediatorSetup: Grafana dashboard updated.

# 3.1

NEW FEATURES:

* OSSMediatorCollector: 
    * Added support for getting CORE alarms.
    * Alarm notification now support DAC and CORE alarms also.
    * Added support to call SIM APIs.
* MediatorSetup: Added dashboards to view ACTIVE/HISTORY CORE alarms.

IMPROVEMENTS:

* OSSMediatorCollector: Changes to write NHG and SIMs data to file.
* ElasticsearchPlugin:
    * Added elasticsearch authentication and old indices deletion feature.
    * Changes for pushing NHG and SIM data to elasticsearch.

# 3.0

NEW FEATURES:

* OSSMediatorCollector: Added support for getting DAC alarms.
* MediatorSetup: Includes setup script and Grafana dashboards used for installation and creation of Grafana dashboards for the PM/FM data collected using OSSMediatorCollector and ElasticsearchPlugin.

IMPROVEMENTS:

* Added Nokia DAC `v2` API support.

REMOVED:

* Removed Nokia DAC `v1` API support.

# 2.6

NEW FEATURES:

* OSSMediatorCollector: Added alarm notifier feature to send notification for configured alarms to MS webhook.
* ElastisearchPlugin: Changes for deleting data collected from elasticsearch based on `elasticsearch_data_retention_duration` field.

# 2.5

BUG FIXES:

* OSSMediatorCollector: API call incorrect end time fix.

# 2.4

NEW FEATURES:

* ElastisearchPlugin: Module for pushing the data collected by OSSMediatorCollector to Elasticsearch.

IMPROVEMENTS:

* OSSMediatorCollector:  
    * Improvements for calling PM/FM APIs network wise.
    * Changes for retrying API call when pagination fails or received number of records is not equal to total number of records.  
    * Added `delay` parameter(in minutes) in config for calling APIs after delay interval.
    * Added `sync_duration` parameter(in minutes) in config, for syncing FM data for the given duration.    
* OpenNMSPlugin: Changes for deleting data collected by collector based on `cleanup_duration`.

# 2.3

IMPROVEMENTS:

* OpenNMSPlugin: Changes for renaming Formatted PM file to use UTC time instead of system local time zone.

# 2.2

IMPROVEMENTS:

* OSSMediatorCollector: Added `password` parameter in conf.json for reading password from base64 encoded string.
* OpenNMSPlugin: Adding `LNCEL` and `LNBTS` ID in formatted data for getting cell level data.

REMOVED:

* OSSMediatorCollector: Removed reading password from console.

# 2.1

IMPROVEMENTS:

* Added Network Hardware Group (NHG) changes.

BUG FIXES:

* Fix panic related to OSSMediatorCollector.
* Added `-v` option to display the version of OSSMediator.
* Changes for clearing alarms.

# 2.0

IMPROVEMENTS:

* Changed OSSMediatorCollector and OpenNMSPlugin to support moss integration with OpenNMS.

REMOVED:

* Removed netAct integration with OpenNMS support from OSSMediator.


# 1.0

NEW FEATURES:

* Added OSSMediatorCollector and OpenNMSPlugin to support netAct integration with OpenNMS.
