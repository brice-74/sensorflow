#!/bin/bash
set -euo pipefail

# initialize dependency scripts
current_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "${SCRIPT_WAITNET_PATH:-"${current_dir}/waitnet.sh"}"
source "${SCRIPT_ARGPARSER_PATH:-"${current_dir}/argparser.sh"}"

# define and parse config with argparser
define_option object-store ${INFLUXDB3_OBJECT_STORE:-"file"}
define_option node-id ${INFLUXDB3_NODE_IDENTIFIER_PREFIX:-"node1"}
define_option log-filter ${LOG_FILTER:-"trace"}
define_option plugin-dir ${INFLUXDB3_PLUGIN_DIR:-"/etc/influxdb/plugins"}
define_option data-dir ${INFLUXDB3_DB_DIR:-"/var/lib/influxdb"}
define_option http-bind ${INFLUXDB3_HTTP_BIND_ADDR:-"0.0.0.0:8181"}
define_option host ${INFLUXDB3_HOST:-"0.0.0.0"}
define_option port ${INFLUXDB3_PORT:-"8181"}
define_option proto ${INFLUXDB3_HTTP_PROTO:-"http"}
define_option token-path "${INFLUXDB3_ADMIN_TOKEN_PATH:-}" required
define_option db-name "${INFLUXDB3_DATABASE_NAME:-}" required
define_option retention "${INFLUXDB3_DATABASE_RETENTION:-}" required

parse_args_and_validate "$@" || exit 1

object_store=$(get_option object-store)
node_id=$(get_option node-id)
log_filter=$(get_option log-filter)
plugin_dir=$(get_option plugin-dir)
data_dir=$(get_option data-dir)
http_bind=$(get_option http-bind)
token_path=$(get_option token-path)
host=$(get_option host)
port=$(get_option port)
proto=$(get_option proto)
db_name=$(get_option db-name)
retention=$(get_option retention)
addr="$proto://$http_bind"

# start influx server
influxdb3 serve \
   --object-store "$object_store" \
   --node-id "$node_id" \
   --log-filter "$log_filter" \
   --plugin-dir "$plugin_dir" \
   --data-dir "$data_dir" \
   --http-bind "$http_bind" &
influx_pid=$!

# wait until influx API is ready
waitnet $host $port 10 1

# create admin token
if [[ ! -f "$token_path" ]]; then
   echo "🔑  Creating admin token..."
   output=$(influxdb3 create token --admin --host "$addr" 2>&1) || {
      echo "[entrypoint]  ❌  Failed to create token. Output:"
      echo "$output"
      exit 1
   }

   token=$(echo "$output" | grep -oP 'apiv3_\S+' | head -n1) || true
   if [[ -z "${token:-}" ]]; then
      echo "[entrypoint]  ❌  Failed to extract token from output:"
      echo "$output"
      exit 1
   fi

   echo "$token" > "$token_path"
   echo "[entrypoint]  ✅  Token created and saved to $token_path"
else
   token=$(cat "$token_path")
   echo "[entrypoint]  ✅  Token already exists at $token_path"
fi

# create database
if influxdb3 show databases --host "$addr" --token "$token" | grep -q "$db_name"; then
   echo "[entrypoint]  ✅  Database '$db_name' already exists."
else
   echo "[entrypoint]  📦  Creating database '$db_name}'..."
   influxdb3 create database "$db_name" \
      --host "$addr" \
      --token "$token" \
      --retention-period "${retention}"
   echo "[entrypoint]  ✅  Database created."
fi

# keep script running after influx serve
wait