#!/bin/bash
set -e

version=$(cat VERSION)
heap_size=${heap_size:-4g}
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

echo "Installing OSSMediator $version version"

#OpenSearch
mkdir -p es_data
chmod g+rwx es_data
chgrp 0 es_data
chown 1000:1000 es_data/
docker start ndac_oss_opensearch 2>/dev/null || echo "Building OpenSearch container..." ; docker run --name "ndac_oss_opensearch" --restart=always -t -d -p 9200:9200 -p 9600:9600 --ulimit nofile=65535:65535 -e "discovery.type=single-node" -e 'DISABLE_SECURITY_PLUGIN=true' -e OPENSEARCH_JAVA_OPTS="-Xms$heap_size -Xmx$heap_size" -v $(pwd)/es_data:/usr/share/opensearch/data opensearchproject/opensearch:2.19.1

#Grafana
mkdir -p grafana_storage
chown 472:472 grafana_storage/
docker start ndac_oss_grafana 2>/dev/null || echo "Building Grafana container..." ; docker run -d --name "ndac_oss_grafana" --restart unless-stopped --network host  -e "GF_INSTALL_PLUGINS=grafana-opensearch-datasource" -v $(pwd)/grafana_storage:/var/lib/grafana -v $(pwd)/grafana_data/provisioning/dashboards:/etc/grafana/provisioning/dashboards -v $(pwd)/grafana_data/provisioning/datasources:/etc/grafana/provisioning/datasources -v $(pwd)/grafana_data/dashboards:/etc/grafana/dashboards grafana/grafana-oss:11.6.1

#Collector DC1
docker start ndac_oss_collector_dc1 2>/dev/null || echo "Building DC1 collector container..." ; docker run -d --name "ndac_oss_collector_dc1" --restart unless-stopped --network host -v $(pwd)/reports/dc1:/reports  -v $(pwd)/collector/logs/dc1:/collector/log -v $(pwd)/collector/checkpoints/dc1:/collector/bin/checkpoints -v $(pwd)/config/dc1/.secret:/collector/bin/.secret -v $(pwd)/config/dc1/collector_conf.json:/collector/resources/conf.json ossmediatorcollector:$version
#sleep 5

#Collector DC4
docker start ndac_oss_collector_dc4 2>/dev/null || echo "Building DC4 collector container..." ; docker run -d --name "ndac_oss_collector_dc4" --restart unless-stopped --network host -v $(pwd)/reports/dc4:/reports  -v $(pwd)/collector/logs/dc4:/collector/log -v $(pwd)/collector/checkpoints/dc4:/collector/bin/checkpoints -v $(pwd)/config/dc4/.secret:/collector/bin/.secret -v $(pwd)/config/dc4/collector_conf.json:/collector/resources/conf.json ossmediatorcollector:$version
#sleep 5

#Collector DC5
docker start ndac_oss_collector_dc5 2>/dev/null || echo "Building DC5 collector container..." ; docker run -d --name "ndac_oss_collector_dc5" --restart unless-stopped --network host -v $(pwd)/reports/dc5:/reports  -v $(pwd)/collector/logs/dc5:/collector/log -v $(pwd)/collector/checkpoints/dc5:/collector/bin/checkpoints -v $(pwd)/config/dc5/.secret:/collector/bin/.secret -v $(pwd)/config/dc5/collector_conf.json:/collector/resources/conf.json ossmediatorcollector:$version
#sleep 5

echo "Building ElasticSearchPlugin container..."
sleep 30

#ElasticSearchPlugin
docker start ndac_oss_elasticsearchplugin 2>/dev/null || docker run -d --name "ndac_oss_elasticsearchplugin" --restart unless-stopped --network host -v $(pwd)/reports/dc1:/reports/dc1 -v $(pwd)/reports/dc4:/reports/dc4 -v $(pwd)/reports/dc5:/reports/dc5 -v $(pwd)/plugin/logs:/plugin/log -v $(pwd)/config/plugin_conf.json:/plugin/resources/conf.json elasticsearchplugin:$version


echo "OSSMediator is started, open http://<IP_Address>:3000/dashboards to view the dashboards."
