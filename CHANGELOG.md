# Changelog
All notable changes to this project will be documented in this file.

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
