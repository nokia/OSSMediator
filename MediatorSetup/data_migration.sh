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

echo "Data migration from elastic search to opensearch"

mkdir -p es_data
mkdir -p es_data/snapshots
chmod g+rwx es_data
chmod g+rwx es_data/snapshots
chgrp 0 es_data
chgrp 0 es_data/snapshots
chown 1000:1000 es_data/
chown 1000:1000 es_data/snapshots/

echo "killing elasticsearch container"
#kill elasticsearch container
elastic_container=$(docker ps -a -q --filter="name=^/ndac_oss_elasticsearch$")

echo "Hello"
echo $elastic_container

docker stop $elastic_container
docker rm $elastic_container

sleep 20

#startup elastic search with repo.path
docker run --name "ndac_oss_elasticsearch" -t -d -p 9200:9200 -p 9300:9300 --ulimit nofile=65535:65535  -e "path.repo=/usr/share/elasticsearch/snapshots" -e "discovery.type=single-node" -e ES_JAVA_OPTS="-Xms$heap_size -Xmx$heap_size -Dlog4j2.formatMsgNoLookups=true"  -v $(pwd)/es_data/snapshots:/usr/share/elasticsearch/snapshots -v $(pwd)/es_data:/usr/share/elasticsearch/data docker.elastic.co/elasticsearch/elasticsearch:7.4.2

echo "registering fs as repo for snapshots"
sleep 60

#register repository as fs
curl -X PUT "localhost:9200/_snapshot/my-fs-repository" -H 'Content-Type: application/json' -d'{"type": "fs", "settings": {"location": "/usr/share/opensearch/snapshots"}}'

echo "take snapshots of all indices"
#take snapshots of all indices
curl -X PUT "localhost:9200/_snapshot/my-fs-repository/1"

#kill elasticsearch docker
elastic_container=$(docker ps -q --filter="name=^/ndac_oss_elasticsearch$")

echo $elastic_container

docker stop $elastic_container
docker rm $elastic_container

sleep 20

echo "Installing opensearch"
#start opensearch
docker run --name "ndac_oss_opensearch" -t -d -p 9200:9200 -p 9600:9600 --ulimit nofile=65535:65535 -e "discovery.type=single-node" -e 'DISABLE_SECURITY_PLUGIN=true' -e OPENSEARCH_JAVA_OPTS="-Xms$heap_size -Xmx$heap_size"  -e "path.repo=/usr/share/opensearch/snapshots" -v $(pwd)/es_data:/usr/share/opensearch/data -v $(pwd)/es_data/snapshots:/usr/share/opensearch/snapshots opensearchproject/opensearch:2.5.0

sleep 60

echo "close all indices before restoring"
#close all indices first before restoring
curl -X POST "localhost:9200/_all/_close"


echo "restore all indices"
#restore all indices
curl -X POST "localhost:9200/_snapshot/my-fs-repository/1/_restore"

echo "data migration completed"