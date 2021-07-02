Installation Steps:
-------------------

Please take the latest OSSMediatorCollector, ElasticsearchPlugin and mediator_startup package files from Info Center - OSS Software Package. 
Contact Support team if packages are not available.

The mediator_startup package is an automated script to bring the setup in few minutes. The script will unpack the contents of the mediator collector and ElasticsearchPlugin, 
install Grafana, Elasticsearch data source. Configures the grafana for the PM KPI, FM Active and History alarm dashboards. To execute this automated script, you need to have 
Docker installed on your system. 

You can find the right version of Docker for your flavour of linux and install them from the following site.
https://docs.docker.com/engine/install/

Once docker is installed on your system:
 
Following steps will setup OSSMediator packages with Elasticsearch data source and Grafana:

1.	Download/Copy the mediator packages from the Info Center into a new directory.
2.	Unzip the mediator_setup.zip in a newly-created directory.
3.	Update collector_config.json template file as described in "Nokia DAC OSS Mediator Collector Configuration" under section "User configuration".
4.	Update plugin_config.json file as described in the "Nokia DAC OSS ElasticSearchPlugin Configuration" under section "User configuration".
5.	Check if execute permissions are there for the script, if not set it as `chmod 777 startup.sh`, then execute `sudo ./startup.sh` command.
6. 	The script will run to do the setup and bringup the dashboards. If there are any errors, please correct them and re-run the script.

The setup should be ready and you can access the PM and FM dashboards from your browser at
`http://<your_Ip_address>:3000/dashboards`.

