#!/bin/bash
set -e

version=$(cat ../VERSION)
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

echo "Installing OSSMediator $version version"

mkdir -p es_data
chmod g+rwx es_data
chgrp 0 es_data
chown 1000:1000 es_data/
docker start ndac_oss_opensearch 2>/dev/null || docker run --name "ndac_oss_opensearch" --restart=always -t -d -p 9200:9200 -p 9600:9600 --ulimit nofile=65535:65535 -e "discovery.type=single-node" -e 'DISABLE_SECURITY_PLUGIN=true' -e OPENSEARCH_JAVA_OPTS="-Xms$heap_size -Xmx$heap_size" -v $(pwd)/es_data:/usr/share/opensearch/data opensearchproject/opensearch:2.13.0

mkdir -p grafana_storage
chown 472:472 grafana_storage/
docker start ndac_grafana 2>/dev/null || docker run -d --name "ndac_grafana" --network host  -e "GF_INSTALL_PLUGINS=grafana-opensearch-datasource" -v $(pwd)/grafana_storage:/var/lib/grafana -v $(pwd)/grafana_data/provisioning/dashboards:/etc/grafana/provisioning/dashboards -v $(pwd)/grafana_data/provisioning/datasources:/etc/grafana/provisioning/datasources -v $(pwd)/grafana_data/dashboards:/etc/grafana/dashboards grafana/grafana-oss:9.5.17

docker start ndac_oss_collector 2>/dev/null || docker run -d --name "ndac_oss_collector" --network host -v $(pwd)/reports:/reports  -v $(pwd)/collector/log:/collector/log -v $(pwd)/collector/checkpoints:/collector/bin/checkpoints -v $(pwd)/.secret:/collector/bin/.secret -v $(pwd)/collector_conf.json:/collector/resources/conf.json ossmediatorcollector:$version
sleep 30

docker start ndac_oss_elasticsearchplugin 2>/dev/null || docker run -d --name "ndac_oss_elasticsearchplugin" --network host -v $(pwd)/reports:/reports -v $(pwd)/plugin/log:/plugin/log -v $(pwd)/plugin_conf.json:/plugin/resources/conf.json elasticsearchplugin:$version

echo "OSSMediator is started, open http://<IP_Address>:3000/dashboards to view the dashboards."
