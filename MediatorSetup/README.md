# Software package installation steps:

The OSS software package contains sub packages for:
* OSSMediatorCollector
* ElasticsearchPlugin
* Grafana dashboard JSON files
* Configuration files
* startup scripts

The startup.sh script is an automated script to bring the setup in few minutes. The script will unpack the contents of the mediator collector and ElasticsearchPlugin, install Grafana, Elasticsearch data source. Configures the grafana for the PM KPI, FM Active and History alarm dashboards.
To execute this automated script, you need to have Docker installed on your system.

You can find the right version of Docker for your flavour of linux and install them from the following site.
https://docs.docker.com/engine/install/

Once the docker is installed on your system:
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

* In case the user’s password is updated, execute `sudo ./storesecret -c collector_conf.json` and input the updated password, then execute `sudo ./startup.sh` to restart the modules again.

* By default, elasticsearch will be started with 2GB heap memory.  
  To increase the heap memory execute `sudo ./startup.sh --heap_size <HEAP_SIZE>`.  
  Ex: `sudo ./startup.sh --heap_size 4g`, `sudo ./startup.sh --heap_size 500m`

# Installation steps with pre-installed elasticsearch:

1. Follow step 1- 5 from the above installation steps.
2. Check if execute permissions are there for the `startup_without_elasticsearch.sh` script, if not set it as `chmod 777 startup_without_elasticsearch.sh`, then execute `sudo ./startup_without_elasticsearch.sh` command.

````
NOTE: Refer "Nokia DAC OSS ElasticSearchPlugin Configuration" document to know more details on indices used for storing data.
````

# Installation steps for running OSSMediator in docker:

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
* ndac_oss_elasticsearch: Elasticsearch instance for storing KPI data.
* ndac_grafana: Grafana instance for dashboards.
* ndac_oss_collector: OSSMediatorCollector for collecting PM/FM data.
* ndac_oss_elasticsearchplugin: ElasticsearchPlugin to push collected data to elasticsearch.

The setup should be ready, and you can access the PM and FM dashboards from your browser at `http://<your_Ip_address>:3000/dashboards`, all the NDAC dashboards will be created inside `Nokia-DAC` directory.  
The Grafana dashboards will appear with login prompt and default credentials are `admin` with password `admin`.

* In case the user’s password is updated, execute `sudo ./storesecret -c collector_conf.json` and input the updated password, then stop and remove `ndac_oss_collector` docker container and execute `sudo ./mediator_docker_startup.sh` to restart the modules again.

* By default, elasticsearch will be started with 2GB heap memory.  
  To increase the heap memory execute `sudo ./mediator_docker_startup.sh --heap_size <HEAP_SIZE>`.  
  Ex: `sudo ./startup.sh --mediator_docker_startup 4g`, `sudo ./mediator_docker_startup.sh --heap_size 500m`

# Steps to modify NDAC dashboards:
1. Make a copy of dashboard and save the dashboard title with prefix other than `NDAC`.
2. Update the new dashboard as per the requirement.

````
NOTE: Please don't update NDAC dashboards, any modifications will be lost after package update.
````