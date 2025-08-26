# OSSMediator Version Upgrade Guide

## Table of Contents
1. [Introduction](#introduction)
2. [Pre-requisites](#pre-requisites)
3. [Backup Existing Data](#backup-existing-data)
4. [Upgrade Steps](#upgrade-steps)
5. [Post-upgrade Verification](#post-upgrade-verification)
6. [Troubleshooting](#troubleshooting)

## Introduction
This document provides step-by-step instructions to upgrade OSSMediator from a previous version to the latest release.

## Pre-requisites
- Ensure you have administrator/sudo access.
- Backup all configuration files and data.
- Review the release notes for configuration changes.

## Backup Existing Data
1. Stop all running OSSMediator services/containers.
   * To stop the service, run:
     `systemctl stop collector`
     `systemctl stop elasticsearchplugin`
     Additionally, stop and remove the `ndac_oss_opensearch` container.
   * To stop Docker containers, run:
     `docker ps -a`, copy the container IDs of `ndac_oss_collector` and `ndac_oss_elasticsearchplugin`, then run: `docker stop <container_id>` and `docker rm <container_id>`
     Additionally stop and remove the `ndac_grafana` and `ndac_oss_opensearch` containers.
2. Backup the following:
    - `collector_conf.json`
    - `plugin_conf.json`
    - Any custom dashboards or scripts.
    - OSSMediatorCollector response directory (e.g., `/reports`)
    - OpenSearch data (`es_data`)
    - Grafana data (`grafana_storage`)
    - Secret directories (`.secret`)

## Upgrade Steps
1. Download the latest OSSMediator package from the Info Center.
2. Unzip and replace the binaries/scripts as per the new package.
3. Update configuration files if there are new fields or changes (refer to the release notes).
4. If using Docker, rebuild and restart containers:
    - `make all`
    - `sudo ./mediator_docker_startup.sh`
5. If using binary installation, run the updated startup scripts.

## Post-upgrade Verification
- Check service/container status.
- Access Grafana dashboards at `http://<your_Ip_address>:3000/dashboards`.
- Verify data collection and dashboard updates.

## Troubleshooting
- Review logs in the `collector/logs` and `plugin/logs` directory.
- To check OpenSearch logs, run: `docker logs ndac_oss_opensearch`.
- Consult the Nokia DAC support for unresolved issues.