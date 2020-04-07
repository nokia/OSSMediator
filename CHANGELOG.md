# Changelog
All notable changes to this project will be documented in this file.

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
