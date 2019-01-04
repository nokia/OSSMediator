# OSSMediator

OSSMediator is used to obtain Performance Management (PM) and Fault Management (FM) data from Nokia DAC and to offer the data for OpenNMS.  
It has following two components:
- OSSMediatorCollector
- OpenNMSPlugin

The OSSMediatorCollector module will call the Nokia NDAC APIs periodically as per the configuration to get the metrics data for the customer’s managed network only.  

The OpenNMSPlugin is a module that will take the data that Mediator collector has collected and convert it into the format needed by the third party Network Management System like OpenNMS in this case.

### Prerequisites

OSSMediator is compatible with only Unix/Linux system.

### Project Structure

    .  
    ├── OSSMediatorCollector        # Source files  
    ├── OpenNMSPlugin               # Source files
    └── README.md  
