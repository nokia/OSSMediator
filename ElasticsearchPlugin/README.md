# ElasticsearchPlugin

The ElasticsearchPlugin is a command line application that monitors the local filesystem, where the PM and FM metrics data is collected by the OSS mediator collector.
Then read PM and FM data is then pushed to elasticsearch data source.

### Prerequisites

ElasticsearchPlugin is compatible with only Unix/Linux system.

### Project Structure
    .  
    ├── resources               # Resource files  
        └── conf.json  
    ├── src                     # Source files  
    ├── Makefile  
    ├── Dockerfile  
    └── README.md  

### Installation steps

ElasticsearchPlugin's binary should be built by running `make all` command followed by `make build_package` command.  
It will create binary named as `elasticsearchplugin` inside `bin` directory and package containing the binary and resource file, named as `ElasticsearchPlugin-<VERSION>.zip` inside `package` directory.  

Please follow below procedure to install ElasticsearchPlugin-<VERSION>.zip in your home directory:

````
$ mkdir plugin
$ cp ElasticsearchPlugin-<VERSION>.zip plugin/
$ cd plugin/
$ unzip ElasticsearchPlugin-<VERSION>.zip
````

ElasticsearchPlugin directory structure after installation will be as shown below:

````
    .
    ├── ElasticsearchPlugin-<VERSION>.zip
    ├── bin
        └── elasticsearchplugin
    ├── log
        └── ElasticsearchPlugin.log
    └── resources
        └── conf.json
````

## Usage
```
Usage: ./elasticsearchplugin [options]
Options:
        -h, --help
                Output a usage message and exit.
        -conf_file
                Config file path (default "../resources/conf.json")
        -log_dir
                Log Directory (default "../log"), logs will be stored in ElasticsearchPlugin.log file.
        -log_level
                Log Level (default 4), logger level in collector.log file. Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)
        -v
                Prints OSSMediator's version
```

## Configuration

ElasticsearchPlugin reads all the collected PM/FM from OSSMediatorCollector and inserts the data to elasticsearch.

* To insert PM/FM metrics in elasticsearch, modify conf.json configuration file under the "resources" directory as shown in the example:

````json
{
  "source_dirs":[
    "<SOURCE DIRECTORY PATH>",
    "<SOURCE DIRECTORY PATH>"
  ],
  "elasticsearch": {
    "url":"http://localhost:9200",
    "user": "<USERNAME>",
    "password": "<PASSWORD>",
    "data_retention_duration": 90
  },
  "cleanup_duration": 1440
}
````

| Field                                 | Type              | Description                                                                                                                                                                                                               |
|---------------------------------------|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| source_dirs                           | [string]          | Base directory path of the respective user where PM/FM data is pushed by the collector. This path has to be same as the path mentioned in response_dest directory of respective user in mediator collector configuration. |
| elasticsearch.url                     | string            | The url to connect to elasticSearch data source. Default: "http://localhost:9200".                                                                                                                                        |
| elasticsearch.user                    | string (Optional) | Elasticsearch user name.                                                                                                                                                                                                  |
| elasticsearch.password                | string (Optional) | Elasticsearch user's password encoded as base64 string.                                                                                                                                                                   |
| elasticsearch.data_retention_duration | integer           | Duration in days, for which ElasticsearchPlugin will cleanup the metrics from Elasticsearch data source.                                                                                                                  |
| cleanup_duration                      | integer           | Duration in minutes, after which ElasticsearchPlugin will cleanup the collected files on the local file system.                                                                                                           |

````
NOTE: If the data collection directories are modified, then the source_dir should be matched with the directory as given in OSSMediatorCollector configuration information.
      The source_dir should be same as the response_dest directory in OSSMediatorCollector configuration.
````

* To start ElasticsearchPlugin, go to the installed path of the mediator bin directory and start by calling the following command:

````
./elasticsearchplugin
````

* ElasticsearchPlugin logs can be checked in ElasticsearchPlugin_HOME/log/ElasticsearchPlugin.log file.


* ElasticsearchPlugin writes data to following indices:

| API                            | Index Pattern               |
|--------------------------------|-----------------------------|
| network-hardware-groups        | nhg-data                    |
| sims                           | sims-data                   |
| account-sims                   | account-sims-data           |
| access-point-sims              | ap-sims-data                |
| pmdata (RADIO) (4G)            | 4g-pm-<METRIC_NAME>-MM-YYYY |
| pmdata (RADIO) (5G)            | 5g-pm-<METRIC_NAME>-MM-YYYY |
| pmdata (EDGE)                  | edge-pm                     |
| pmdata (CORE)                  | core-pm                     |
| fmdata (RADIO) (ACTIVE)        | radio-fm                    |
| fmdata (RADIO) (HISTORY)       | radio-fm                    |
| fmdata (CORE) (ACTIVE)         | core-fm                     |
| fmdata (CORE) (HISTORY)        | core-fm                     |
| fmdata (DAC) (ACTIVE)          | dac-fm                      |
| fmdata (DAC) (HISTORY)         | dac-fm                      |
| fmdata (APPLICATION) (ACTIVE)  | application-fm              |
| fmdata (APPLICATION) (HISTORY) | application-fm              |

`MM-YYYY` denotes month and years of the collected metric time.

````
NOTE: Elasticsearchplugin deletes data from the above mentioned indices as per the configuration provided.  
      Please be careful when using the plugin with pre-installed elasticsearch instance, as it will delete other data if index pattern is similar.
````
