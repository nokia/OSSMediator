# OSS Mediator Collector

The OSS Mediator Collector is a command line API client. It connects with NDAC APIGW and gets the FM and PM data using REST interface at regular intervals based on the collector configuration for the customer’s managed networks only.

### Prerequisites

MediatorCollector is compatible with only Unix/Linux system.

### Project Structure

    .  
    ├── resources               # Resource files  
        └── resource.conf  
    ├── src                     # Source files  
    ├── Makefile  
    ├── Dockerfile  
    └── README.md  

### Installation steps

OSSMediatorCollector's binary should be built by running `make all` command.  
It will create binary named as `collector` inside `bin` directory and package containing the binary and resource file, named as `OSSMediatorCollector.zip` inside `package` directory.  
  
Please follow below procedure to install OSSMediatorCollector.zip in your home directory:

````
$ mkdir collector
$ cp OSSMediatorCollector.zip collector/
$ cd collector/
$ unzip OSSMediatorCollector.zip
````

MediatorCollector directory structure after installation will be as shown below:

````
    .
    ├── OSSMediatorCollector.zip
    ├── bin
        └── collector
    ├── log
        └── collector.log
    └── resources
        └── resource.conf
````

## Usage

Usage: ./collector [options]  
Options:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-h, --help  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Output a usage message and exit.  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-conf_file string  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Config file path (default "../resources/conf.json")  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-cert_file string  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Certificate file path (if cert_file is not passed then it will establish TLS auth using root certificates.)  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-log_dir string  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Log Directory (default "../log"), logs will be stored in collector.log file.  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-log_level int  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Log Level (default 4), logger level in collector.log file. Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-tls_skip  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Skip TLS Authentication  

## Configuration

PM / FM data collection by collector is performed using REST interface at regular intervals based on configuration.  

* To collect PM / FM data, it is required to modify conf.json configuration file in "resource" directory as shown in the example.

````json
{
    "base_url": "https://api.dac.nokia.com/api/v1",
    "users": [
        {
            "email_id": "<USER EMAIL>",
            "response_dest": "<DIRECTORY PATH>"
        },
        {
            "email_id": "<USER EMAIL>",
            "response_dest": "<DIRECTORY PATH>"
        }
    ],
    "um_api": {
        "login": "/sessions/login",
        "refresh": "/sessions/refresh",
        "logout": "/sessions/logout"
    },
    "apis": [
        {
            "api": "/metrics/radios/pmdata",
            "interval": 15
        },
        {
            "api": "/metrics/radios/fmdata",
            "type": "ACTIVE",
            "interval": 60
        },
        {
            "api": "/metrics/radios/fmdata",
            "type": "HISTORY",
            "interval": 15
        }
    ],
    "limit": 100
}
````

| Field          | Type        | Description                                                                                                                                                                   |
|----------------|-------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| base_url       | string      | APIGW base URL.                                                                                                                                                               |
| users          | [user_conf] | Users details.                                                                                                                                                                |
| email_id       | string      | User's email ID.                                                                                                                                                              |
| response_dest  | string      | Base directory to store the response from the REST APIs. Subdirectories will be created inside the base directory for storing each APIs response in their respective location |
| um_api         | object      | User management APIs.                                                                                                                                                         |
| um_api.login   | string      | Customer portal login API.                                                                                                                                                    |
| um_api.refresh | string      | Customer portal refresh session API.                                                                                                                                          |
| um_api.logout  | string      | Customer portal logout API.                                                                                                                                                   |
| apis           | [api_conf]  | Get PM/FM APIs.                                                                                                                                                               |
| api            | string      | API URL of get PM/FM data.                                                                                                                                                    |
| interval       | integer     | Interval at which API should be called to collect data.                                                                                                                       |
| type           | string      | Type of FM request ("ACTIVE" or "HISTORY")                                                                                                                                    |
| limit          | integer     | Number of records to be fetched from the API.

* To start collector, go to the installed path of the collector bin directory and start by calling the following command:

````
./collector
````

Enter the password for each customer having the right permission.  
NOTE: For login details (email ID and password) contact Nokia DAC support/operations team.  

Once the login is successful for all users, the collector will periodically start collecting the data by calling the configured APIs for the customer’s managed network.

Collector logs can be checked in $cd $collector_basepath/log/collector.log file.
