# Changelog
All notable changes to this project will be documented in this file.

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
