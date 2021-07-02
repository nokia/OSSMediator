# OSSMediator

[![Build Status](https://travis-ci.org/nokia/OSSMediator.svg?branch=master)](https://travis-ci.org/nokia/OSSMediator)

### Overview

OSSMediator is used to obtain Performance Management (PM) and Fault Management (FM) data from Nokia DAC and to offer the data for any 3rd party NMS system like OpenNMS or any external datasource like Elasticsearch.  
It has following three components:
- OSSMediatorCollector
- OpenNMSPlugin (**deprecated**)
- ElasticsearchPlugin
- MediatorSetup

The OSSMediatorCollector module will call the Nokia DAC APIs periodically as per the configuration to get the metrics data for the customer’s managed network only.  

The OpenNMSPlugin is a module that will take the data that Mediator collector has collected and convert it into the format needed by the third party Network Management System like OpenNMS in this case.

The ElasticsearchPlugin reads data collected by Mediator collector and pushes it to elasticsearch.

MediatorSetup includes shell script and Grafana dashboards used for installation and creation of Grafana dashboards for the PM/FM data collected using OSSMediatorCollector and ElasticsearchPlugin.  

### Prerequisites

OSSMediator is compatible with only Unix/Linux system.

### Project Structure

    .  
    ├── OSSMediatorCollector        # Source files  
    ├── OpenNMSPlugin               # Source files
    ├── ElasticsearchPlugin         # Source files
    ├── MediatorSetup              # Source files
    └── README.md  


## License

This project is licensed under the BSD-3-Clause license - see the [LICENSE](https://github.com/nokia/OSSMediator/blob/master/LICENSE).