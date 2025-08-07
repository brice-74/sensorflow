#!/bin/bash
set -euo pipefail

current_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "${ARGPARSER_PATH:-"${current_dir}/argparser.sh"}"

define_option object-store ${INFLUXDB3_OBJECT_STORE:-"file"}
define_option node-id ${INFLUXDB3_NODE_IDENTIFIER_PREFIX:-"node1"}
define_option log-filter ${LOG_FILTER:-"trace"}
define_option plugin-dir ${INFLUXDB3_PLUGIN_DIR:-"/etc/influxdb/plugins"}
define_option data-dir ${INFLUXDB3_DB_DIR:-"/var/lib/influxdb"}
define_option http-bind ${INFLUXDB3_HTTP_BIND_ADDR:-"0.0.0.0:8181"}

parse_args "$@"

influxdb3 serve \
   --object-store  $(get_option object-store) \
   --node-id  $(get_option node-id) \
   --log-filter $(get_option log-filter) \
   --plugin-dir $(get_option plugin-dir) \
   --data-dir $(get_option data-dir) \
   --http-bind $(get_option http-bind) &
influx_pid=$!

"${current_dir}/wait.sh"
"${current_dir}/create-token-admin.sh"
"${current_dir}/create-database.sh"

wait 