# Table of contents
1. [Introduction](#introduction)
2. [OSSMediator Installation](#installation)
   1. [System Requirements](#system_req)
   2. [Installation with binary package](#package_install)
   3. [Installation with pre-installed elasticsearch/OpenSearch](#pre_elk_install)
   4. [Installation with Docker](#docker_install)
3. [Steps to migrate from Elasticsearch to OpenSearch](#elk_to_opensearch)
4. [Steps to modify NDAC dashboards](#dashboard_update)

## Introduction <a name="introduction"></a>

MediatorSetup includes scripts and Grafana template dashboards and datasource files used for installation and creation of Grafana dashboards for the PM/FM data collected using OSSMediatorCollector and ElasticsearchPlugin.  
MediatorSetup contents:

| Directory/file Name              | Description                                                                                             |
|----------------------------------|---------------------------------------------------------------------------------------------------------|
| grafana_data                     | Contains template Grafana dashboards and provisioned datasources by NDAC.                               |
| service_template                 | Template files for running OSSMediatorCollector and ElasticsearchPlugin as a service.                   |
| data_migration.sh                | Script to migrate data from Elasticsearh to OpenSearch.                                                 |
| grafana_cleanup.sh               | Script to cleanup NDAC provided Grafana datasources and dashboards.                                     |
| mediator_docker_startup.sh       | OSSMediator startup script to run inside docker.                                                        |
| startup.sh                       | OSSMediator startup script to start from binary package.                                                |
| startup_without_elasticsearch.sh | OSSMediator startup script to start from binary package with pre-installed Elasticsearch or OpenSearch. |
| collector_conf.json              | OSSMediatorCollector configuration file, default configuration file with the basic FM/PM APIs.          |
| collector_conf_all_api.json      | OSSMediatorCollector configuration file, it has SIM API added along with the basic FM/PM APIs.          |
| plugin_conf.json                 | ElasticsearchPlugin configuration file.                                                                 |

## OSSMediator Installation <a name="installation"></a>

````
NOTE: OSSMediator's 3.7 version onwards it will install OpenSearch to store the KPI data.
      If user wants to use elasticsearch please install elasticsearch (version > 7.10) and remove OpenSearch installation step from the startup script.
````

### System Requirements <a name="system_req"></a>

* Docker: You can find the right version of Docker for your flavour of linux and install them from the following site.
  https://docs.docker.com/engine/install/
  (Not needed for Installation with pre-installed Elasticsearch/OpenSearch with binary package).
*

### Installation with binary package <a name="package_install"></a>

Follow the below steps to setup OSSMediator on your instance:
1. Download/Copy the mediator packages from the Info Center by clicking "Download Package" button into a new directory.
2. Unzip the OSSPackage-xx.zip in the new directory.
3. Update collector_conf.json template with the correct "base_url", "users" section with the credentials  ("email_id") provided by NDAC Operations/Helpdesk team.  
   Update the "response_dest" with the directory path where you want the collected metrics data to be stored locally on the disk.  
   You can refer the document "Nokia DAC OSS Mediator Collector Configuration" under section "User configuration" for clarification.  
   In the OSS Mediator package there are 2 template files provided, the default "collector_conf.json file" and "collector_conf_all_api.json" file. The "collector_conf_all_api.json" template file has SIM API added along with the basic FM/PM APIs.  
   If you want to collect SIM, APPLICATION metrics as well, then first copy "collector_conf_all_api.json" into "collector_conf.json" file and then edit the file as described above.
4. Update plugin_conf.json template file for the field `<SOURCE DIRECTORY PATH>` with the same directory path as set in above step.
   You can refer to the document "Nokia DAC OSS ElasticSearchPlugin Configuration" under section "User configuration" for clarification.
5. Configure the passwords for the users configured in collector_conf.json file for API access by executing `storesecret`.  
   Check if execute permissions are there for the `storesecret` binary, if not set it as `chmod 777 storesecret`, then execute `sudo ./storesecret -c collector_conf.json` command to store the user passwords.
6. Use `grafana_cleanup.sh` script to clean grafana, check if execute permissions are there for the `grafana_cleanup.sh` script, if not set it as `chmod 777 grafana_cleanup.sh`.  
   `sudo ./grafana_cleanup.sh`.
7. Check if execute permissions are there for the `startup.sh` script, if not set it as `chmod 777 startup.sh`, then execute `sudo ./startup.sh` command.
8. The script will run to do the setup and bring up the dashboards. If there are any errors, please correct them and re-run the script.

The setup should be ready, and you can access the PM and FM dashboards from your browser at `http://<your_Ip_address>:3000/dashboards`, all the NDAC dashboards will be created inside `Nokia-DAC` directory.  
The Grafana dashboards will appear with login prompt and default credentials are `admin` with password `admin`.

* To update the user’s password, execute `sudo ./storesecret -c collector_conf.json` and input the updated password, then execute `sudo ./startup.sh` to restart the modules again.

* By default, OpenSearch will be started with 2GB heap memory.  
  To increase the heap memory execute `sudo ./startup.sh --heap_size <HEAP_SIZE>`.  
  Ex:
    * `To start OpenSearch with 4GB heap memory: sudo ./startup.sh --heap_size 4g`
    * `To start OpenSearch with 500MB heap memory: sudo ./startup.sh --heap_size 500m`

### Installation with pre-installed elasticsearch/OpenSearch <a name="pre_elk_install"></a>

1. Follow step 1- 5 from the above installation steps.
2. Check if execute permissions are there for the `startup_without_elasticsearch.sh` script, if not set it as `chmod 777 startup_without_elasticsearch.sh`, then execute `sudo ./startup_without_elasticsearch.sh` command.

* To update the user’s password, execute `sudo ./storesecret -c collector_conf.json` and input the updated password, then execute `sudo ./startup_without_elasticsearch.sh` to restart the modules again.

````
NOTE: Refer "Nokia DAC OSS ElasticSearchPlugin Configuration" document to know more details on indices used for storing data.
````

### Installation with Docker <a name="docker_install"></a>

1. Move to `OSSMediator` base directory and run `make all` command. It will create two docker images `ossmediatorcollector:<VERSION>` and `elasticsearchplugin:<VERSION>`.
2. Run `cd MediatorSetup` command.
3. Update collector_conf.json template with the correct "base_url", "users" section with the credentials ("email_id") provided by NDAC Operations/Helpdesk team.  
   Update the `response_dest` with `/reports` value.  
   You can refer the document "Nokia DAC OSS Mediator Collector Configuration" under section "User configuration" for clarification.  
   In the OSS Mediator package there are 2 template files provided, the default "collector_conf.json file" and "collector_conf_all_api.json" file. The "collector_conf_all_api.json" template file has SIM API added along with the basic FM/PM APIs.  
   If you want to collect SIM, APPLICATION metrics as well, then first copy "collector_conf_all_api.json" into "collector_conf.json" file and then edit the file as described above.
4. Update plugin_conf.json template file for the field `<SOURCE DIRECTORY PATH>` with `/reports` value (Remove duplicate values).  
   You can refer to the document "Nokia DAC OSS ElasticSearchPlugin Configuration" under section "User configuration" for clarification.
5. Configure the passwords for the users configured in collector_conf.json file for API access by executing `storesecret`.  
   Check if execute permissions are there for the `storesecret` binary, if not set it as `chmod 777 storesecret`, then execute `sudo ./storesecret -c collector_conf.json` command to store the user passwords.
6. Check if execute permissions are there for the `mediator_docker_startup.sh` script, then execute `sudo ./mediator_docker_startup.sh` script to start the OSSMediator modules inside docker.

`mediator_docker_startup.sh` script will create following docker containers:
* ndac_oss_opensearch: OpenSearch instance for storing KPI data.
* ndac_grafana: Grafana instance for dashboards.
* ndac_oss_collector: OSSMediatorCollector for collecting PM/FM data.
* ndac_oss_elasticsearchplugin: ElasticsearchPlugin to push collected data to OpenSearch.

The setup should be ready, and you can access the PM and FM dashboards from your browser at `http://<your_Ip_address>:3000/dashboards`, all the NDAC dashboards will be created inside `Nokia-DAC` directory.  
The Grafana dashboards will appear with login prompt and default credentials are `admin` with password `admin`.

* To update the user’s password, execute `sudo ./storesecret -c collector_conf.json` and input the updated password, then stop and remove `ndac_oss_collector` docker container and execute `sudo ./mediator_docker_startup.sh` to restart the modules again.

* By default, OpenSearch will be started with 2GB heap memory.  
  To increase the heap memory execute `sudo ./mediator_docker_startup.sh --heap_size <HEAP_SIZE>`.  
  Ex:
    * `To start OpenSearch with 4GB heap memory: sudo ./mediator_docker_startup.sh --heap_size 4g`
    * `To start OpenSearch with 500MB heap memory: sudo ./mediator_docker_startup.sh --heap_size 500m`

## Steps to migrate from Elasticsearch to OpenSearch <a name="elk_to_opensearch"></a>

There are various approaches to migrate data from Elasticsearch to OpenSearch. Please refer `https://opensearch.org/docs/latest/upgrade-to/index/`.  
For our MediatorSetup, we opted for taking snapshots of the running Elasticsearch indices, create an OpenSearch cluster and restore the snapshots on the new cluster.  
Execute `sudo ./data_migration.sh` to migrate data from Elasticsearch to OpenSearch.  
`data_migration.sh` script has the following steps:
1. Kill the running elasticsearch container and restart elasticsearch with the env `path.repo`. This configuration allows  to provide a shared filesystem repository for storing the snapshots.
2. After the new elasticsearch container is up and running, register the shared filesystem(specified in the path.repo conf) using elasticsearch rest API.
3. Take snapshots of all the indices
4. Kill the elasticsearch container and install OpenSearch with the same repo.path env configuration.
5. On startup, the indices are automatically restored but are then closed. So, after restoring, open all the indices.

## Steps to modify NDAC dashboards <a name="dashboard_update"></a>

1. Make a copy of dashboard and save the dashboard title with prefix other than `NDAC`.
2. Update the new dashboard as per the requirement.

````
NOTE: Please don't update NDAC dashboards, any modifications will be lost after package update.
````
