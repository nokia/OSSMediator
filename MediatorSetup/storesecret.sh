#!/bin/bash
set -e

secretDir=.secret
mkdir -p $secretDir
chmod 600 $secretDir
for user in $(jq -r '.users[]|"\(.email_id)"' collector_conf.json); do
    read -r -s -p "Enter password for $user:" password
    secretFile="$secretDir/.$user"
    echo -n $password | base64 > $secretFile
    chmod 600 $secretFile
    echo -e "\nSecret stored for $user"
done
