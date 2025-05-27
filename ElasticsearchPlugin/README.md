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
                Log Level (default 4). Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)
        -v
                Prints OSSMediator's version
```

## Configuration

ElasticsearchPlugin reads all the collected PM/FM from OSSMediatorCollector and inserts the data to Elasticsearch/OpenSearch.

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
    "data_retention_duration": 90,
    "initialize_cluster_setting": true,
    "max_shards_per_node": 2000
  },
  "cleanup_duration": 60,
  "max_concurrent_process": 1
}
````

| Field                                    | Type               | Description                                                                                                                                                                                                                                                                                                                         |
|------------------------------------------|--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| source_dirs                              | [string]           | Base directory path of the respective user where PM/FM data is pushed by the collector. This path has to be same as the path mentioned in response_dest directory of respective user in mediator collector configuration.                                                                                                           |
| elasticsearch.url                        | string             | The url to connect to Elasticsearch/OpenSearch data source. Default: "http://localhost:9200".                                                                                                                                                                                                                                       |
| elasticsearch.user                       | string (Optional)  | Elasticsearch/OpenSearch user name.                                                                                                                                                                                                                                                                                                 |
| elasticsearch.password                   | string (Optional)  | Elasticsearch/OpenSearch user's password encoded as base64 string.                                                                                                                                                                                                                                                                  |
| elasticsearch.data_retention_duration    | integer            | Duration in days, for which ElasticsearchPlugin will cleanup the metrics from Elasticsearch/OpenSearch data source. Default value is 90 days.                                                                                                                                                                                       |
| elasticsearch.initialize_cluster_setting | bool               | Default value is true. Enable setting Elasticsearch/OpenSearch configuration for Nokia DAC KPIs (index templates, max. no. of shards, search.max_buckets and script.max_compilations_rate).                                                                                                                                         |
| elasticsearch.max_shards_per_node        | integer            | Maximum shards in Elasticsearch/OpenSearch, default value is 2000. Only applicable if `elasticsearch.set_default_setting` is set to `true`. OpenSearch requires two shards for each indices created for NDAC KPIs, OSSMediator creates ~250 indices monthly. Keep the no. of shards as per `elasticsearch.data_retention_duration`. |
| cleanup_duration                         | integer            | Duration in minutes, after which ElasticsearchPlugin will cleanup the collected files on the local file system. Default value is 60m.                                                                                                                                                                                               |
| max_concurrent_process                   | integer (Optional) | Default value is 1. Maximum no. of concurrent process for pushing PM/FM data to Elasticsearch/OpenSearch.                                                                                                                                                                                                                           |

````
NOTE: 
    * If the data collection directories are modified, then the source_dir should be matched with the directory as given in OSSMediatorCollector configuration information.
      The source_dir should be same as the response_dest directory in OSSMediatorCollector configuration.

    * Keeping high value for max_concurrent_process will increase the CPU/memory usage by ElasticsearchPlugin.
      It is advised to set this value depending on the network size.
      ex: Small Network (1 nhg) : 1
          Medium Network (5 nhg) : 10
          Large Network (20 nhg) : 50
          XL Network (50 nhg) : 100
          XXL Network (150 nhg) : 200

    * Setting elasticsearch.initialize_cluster_setting to true will set following setting in OpenSearch:
        * Enable index deletion with wildcard to cleanup old indices.
          curl -XPUT "<OpenSearch_URL>/_cluster/settings" -H 'Content-Type: application/json' -d'
          {
            "transient": {
              "action.destructive_requires_name": false
            }
          }'
        * Create index template for Nokia DAC KPIs.
          curl -XPUT "<OpenSearch_URL>/_index_template/dac-index" -H 'Content-Type: application/json' -d'
          {
            "index_patterns": [
              "4g-pm*", "5g-pm*", "ixr-pm", "edge-pm", "core-pm", "radio-fm", "dac-fm", "core-fm", "application-fm", "ixr-fm", "nhg-data", "sims-data", "account-sims-data", "ap-sims-data"
            ],
            "template": {
              "settings": {
                "index.number_of_shards": 1
              }
            }
          }'
        * Set cluster settings (cluster.max_shards_per_node, search.max_buckets and script.max_compilations_rate).
          curl -XPUT "<OpenSearch_URL>/_cluster/settings" -H 'Content-Type: application/json' -d'
          {
            "persistent": {
              "cluster.max_shards_per_node": "MAX_SHARDS",
              "search.max_buckets": "100000",
              "script.max_compilations_rate": "2000/5m"
            }
          }'

      If elasticsearch.initialize_cluster_setting is set to false, it is recommended to set the above mentioned setting on OpenSearch.
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
| pmdata (IXR)                   | ixr-pm                      |
| fmdata (RADIO) (ACTIVE)        | radio-fm                    |
| fmdata (RADIO) (HISTORY)       | radio-fm                    |
| fmdata (CORE) (ACTIVE)         | core-fm                     |
| fmdata (CORE) (HISTORY)        | core-fm                     |
| fmdata (DAC) (ACTIVE)          | dac-fm                      |
| fmdata (DAC) (HISTORY)         | dac-fm                      |
| fmdata (APPLICATION) (ACTIVE)  | application-fm              |
| fmdata (APPLICATION) (HISTORY) | application-fm              |
| fmdata (IXR) (ACTIVE)          | ixr-fm                      |
| fmdata (IXR) (HISTORY)         | ixr-fm                      |

`MM-YYYY` denotes month and years of the collected metric time.

````
NOTE: Elasticsearchplugin deletes data from the above mentioned indices as per the configuration provided.  
      Please be careful when using the plugin with pre-installed elasticsearch instance, as it will delete other data if index pattern is similar.
````
