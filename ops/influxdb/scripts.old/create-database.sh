#!/bin/bash
set -euo pipefail

source "${ARGPARSER_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/argparser.sh}"

define_option token-path "${INFLUXDB3_ADMIN_TOKEN_PATH:-"/var/lib/influxdb/admin.token"}"
define_option addr "http://${INFLUXDB3_HTTP_BIND_ADDR:-"0.0.0.0:8181"}"
define_option db-name "${INFLUXDB3_DATABASE_NAME}" required
define_option retention "${INFLUXDB3_RETENTION}" required

parse_args_and_validate "$@"

token_file="$(get_option token-path)"
addr="$(get_option addr)"
db_name="$(get_option db-name)"
retention="$(get_option retention)"

echo "📦  Checking if database '${INFLUXDB3_DATABASE_NAME}' exists..."

if influxdb3 show databases --host "$addr" --token "$token" | grep -q "${INFLUXDB3_DATABASE_NAME}"; then
   echo "✅  Database '${INFLUXDB3_DATABASE_NAME}' already exists."
else
   echo "📦  Creating database '${INFLUXDB3_DATABASE_NAME}'..."
   influxdb3 create database "${INFLUXDB3_DATABASE_NAME}" \
      --host "$addr" \
      --token "$token" \
      --retention-period "${INFLUXDB3_RETENTION}"
   echo "✅  Database created."
fi