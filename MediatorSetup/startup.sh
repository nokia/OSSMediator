#!/bin/bash
set -e

install_grafana() {
	if [ -x "$(command -v yum)" ]; then
		yum install -y https://dl.grafana.com/oss/release/grafana-9.5.15-1.x86_64.rpm
	elif [ -x "$(command -v apt-get)" ]; then
		apt-get install -y adduser libfontconfig1 musl
		wget https://dl.grafana.com/oss/release/grafana_9.5.15_amd64.deb
		dpkg -i grafana_9.5.15_amd64.deb
	elif [ -x "$(command -v rpm)" ]; then
		wget https://dl.grafana.com/oss/release/grafana-9.5.15-1.x86_64.rpm
		rpm -Uvh grafana-9.5.15-1.x86_64.rpm
	else
		echo "Error can't install Grafana, please install it manually and re-run the script."
		exit 1;
	fi
}

function version_gt() {
  test "$(printf '%s\n' "$@" | sort -V | head -n1)" = "$1";
}

heap_size=${heap_size:-2g}
while [ $# -gt 0 ]; do

   if [[ $1 == *"--"* ]]; then
        param="${1/--/}"
        declare $param="$2"
   fi

  shift
done

if ! [ -x "$(command -v docker)" ]; then
  echo "Docker not installed, please install docker and re-run the script."
	exit 1
fi

if ! [ -x "$(command -v grafana-server)" ]; then
  echo "Installing Grafana..."
	install_grafana
	echo "Grafana version is $(grafana-server -v)."
else
  grafana_version=$(grafana-server -v | cut -d' ' -f 2)
  min_grafana_version=9.5.13

  if version_gt $min_grafana_version $grafana_version; then
    echo "Grafana version is $grafana_version."
  else
    echo "Grafana version $grafana_version is less than recommended version $min_grafana_version, updating grafana..."
    install_grafana
    echo "Grafana version is $(grafana-server -v)."
  fi
fi

echo "Installing OpenSearch"
mkdir -p es_data
chmod g+rwx es_data
chgrp 0 es_data
chown 1000:1000 es_data/

echo "Installing OSSMediator"
mkdir -p collector plugin
unzip -o OSSMediatorCollector*.zip -d ./collector
cp collector_conf.json ./collector/resources/conf.json
cp -r .secret/ ./collector/bin/

unzip -o ElasticsearchPlugin*.zip -d ./plugin
cp plugin_conf.json ./plugin/resources/conf.json

echo "Creating services file"
sed -i -e 's?COLLECTOR_PATH?'`pwd`\/collector'?g' ./services_template/collector.service
sed -i -e 's?PLUGIN_PATH?'`pwd`\/plugin'?g' ./services_template/elasticsearchplugin.service
cp ./services_template/collector.service /etc/systemd/system/.
cp ./services_template/elasticsearchplugin.service /etc/systemd/system/.

if [ -x "$(command -v sestatus)" ] && [ $(sestatus | cut -d':' -f 2 | awk '{print $1}' | head -1) == "enabled" ] ; then
	echo "Disabling selinux"
  chcon -h system_u:object_r:bin_t:s0 ./collector/bin/collector
	chcon -h system_u:object_r:bin_t:s0 ./plugin/bin/elasticsearchplugin
fi

echo "Setting up Grafana"
mkdir -p /etc/grafana/dashboards /etc/grafana/provisioning/datasources /etc/grafana/provisioning/dashboards
cp ./grafana_data/provisioning/dashboards/dac_oss_dashboards.yaml /etc/grafana/provisioning/dashboards/.
cp ./grafana_data/provisioning/datasources/dac_oss_datasources.yaml /etc/grafana/provisioning/datasources/.
cp ./grafana_data/dashboards/*.json /etc/grafana/dashboards/.
chmod 775 /etc/grafana/provisioning/datasources/*
chmod 775 /etc/grafana/provisioning/dashboards/*
chmod 775 /etc/grafana/dashboards/*
grafana-cli plugins install grafana-opensearch-datasource

systemctl daemon-reload
systemctl enable collector.service
systemctl enable elasticsearchplugin.service
systemctl enable grafana-server.service

echo "Starting OpenSearch"
name='ndac_oss_opensearch'
if [[ $(docker ps -f "name=$name" --format '{{.Names}}') == $name ]]; then
  docker update --restart=always $name
else
  docker run --name "$name" --restart=always -t -d -p 9200:9200 -p 9600:9600 --ulimit nofile=65535:65535 -e "discovery.type=single-node" -e 'DISABLE_SECURITY_PLUGIN=true' -e OPENSEARCH_JAVA_OPTS="-Xms$heap_size -Xmx$heap_size" -v $(pwd)/es_data:/usr/share/opensearch/data opensearchproject/opensearch:2.9.0
fi

echo "Checking OpenSearch status"
Status=`docker inspect --format "{{.State.Running}}" $name` || true
if [ "$Status" == "true" ]; then
    echo "OpenSearch service started successfully"
else
	echo 'OpenSearch failed See docker ps -a --filter "name=ndac_oss_opensearch" and docker logs ndac_oss_opensearch for details.'
  exit 1
fi

echo "Restarting Grafana "
systemctl restart --quiet grafana-server
sleep 10

echo "Checking Grafana status"
Status=`systemctl is-active grafana-server` || true
if [ "$Status" == "active" ]; then
    echo "Grafana service started successfully"
else
	echo 'Job for grafana-server.service failed See "systemctl status grafana-server.service" and "/var/log/grafana/grafana.log" for details.'
  exit 1
fi

systemctl restart collector
sleep 10
systemctl restart elasticsearchplugin

echo "Checking OSSMediatorCollector status"
Status=`systemctl is-active collector` || true
if [ "$Status" == "active" ]; then
  echo "OSSMediatorCollector service started successfully"
else
	echo 'OSSMediatorCollector failed See "systemctl status collector" and "./collector/log/collector.log" for details.'
  exit 1
fi

echo "Checking ElasticsearchPlugin status"
Status=`systemctl is-active elasticsearchplugin` || true
if [ "$Status" == "active" ]; then
  echo "ElasticsearchPlugin service started successfully"
else
	echo 'ElasticsearchPlugin failed See "systemctl status elasticsearchplugin" and "./plugin/log/ElasticsearchPlugin.log" for details.'
  exit 1
fi

echo "OSSMediator is started, open http://<IP_Address>:3000/dashboards to view the dashboards."
