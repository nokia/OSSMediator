#!/bin/bash
set -e

if ! [ -x "$(command -v sqlite3)" ]; then
  echo "Installing sqlite3..."
  apt-get install -y sqlite3
fi

systemctl stop grafana-server
rm -rf /etc/grafana/provisioning/datasources/oss_datasources.yaml /etc/grafana/provisioning/dashboards/oss_dashboards.yaml /var/lib/grafana/dashboards/4g_core_pm_dashboard.json /var/lib/grafana/dashboards/dac_fm_active_dashboard.json /var/lib/grafana/dashboards/4g_pm_dashboard.json /var/lib/grafana/dashboards/dac_fm_history_dashboard.json /var/lib/grafana/dashboards/4g_pm_kpi_reporting_dashboard.json /var/lib/grafana/dashboards/edge_pm_dashboard.json /var/lib/grafana/dashboards/4g_radio_multefire_pm_accessibility_and_retainability_dashboard.json /var/lib/grafana/dashboards/4g_radio_multefire_pm_system_program_dashboard.json /var/lib/grafana/dashboards/4g_radio_system_program_report.json /var/lib/grafana/dashboards/radio_fm_active_dashboard.json /var/lib/grafana/dashboards/accessibility_5G002.json /var/lib/grafana/dashboards/radio_fm_history_dashboard.json /var/lib/grafana/dashboards/application_fm_active_dashboard.json /var/lib/grafana/dashboards/retainability_5G003.json /var/lib/grafana/dashboards/application_fm_history_dashboard.json /var/lib/grafana/dashboards/slice_performance_5G035.json /var/lib/grafana/dashboards/core_fm_active_dashboard.json /var/lib/grafana/dashboards/system_program_nrcell_level_5G001.json /var/lib/grafana/dashboards/core_fm_history_dashboard.json /var/lib/grafana/dashboards/accessibility-5G002-RS5G_5G19B.json /var/lib/grafana/dashboards/pm_dashboard.json /var/lib/grafana/dashboards/pm_kpi_reporting_dashboard.json /var/lib/grafana/dashboards/retainability-5G003-RS5G_5G19B.json /var/lib/grafana/dashboards/system-program-nrcell-level-5G001-RS5G_5G19B.json
rm -rf /etc/grafana/provisioning/datasources/dac* /etc/grafana/provisioning/dashboards/dac* /var/lib/grafana/dashboards/dac*
echo "delete from dashboard where title like 'NDAC%';" | sqlite3 /var/lib/grafana/grafana.db
echo "delete from dashboard_provisioning where name like 'OSS%';" | sqlite3 /var/lib/grafana/grafana.db
echo "delete from data_source where name like 'dac-%';" | sqlite3 /var/lib/grafana/grafana.db
echo "delete from data_source where name like 'radio-%';" | sqlite3 /var/lib/grafana/grafana.db
echo "delete from data_source where name like 'core-%';" | sqlite3 /var/lib/grafana/grafana.db
echo "delete from data_source where name='application-fm';" | sqlite3 /var/lib/grafana/grafana.db
echo "delete from data_source where name='dac-fm';" | sqlite3 /var/lib/grafana/grafana.db
echo "delete from data_source where name='edge-pm';" | sqlite3 /var/lib/grafana/grafana.db
