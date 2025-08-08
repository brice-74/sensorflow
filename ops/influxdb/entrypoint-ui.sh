#!/bin/sh
set -eo pipefail

# initialize dependency scripts
. "${SCRIPT_ARGPARSER_PATH:-"$(cd "$(dirname "$0")" && pwd)argparser.sh"}"

# define and parse config with argparser
define_option token-path "${INFLUXDB3_ADMIN_TOKEN_PATH}" required
define_option db-name "${INFLUXDB3_DATABASE_NAME}" required
define_option influx-addr ${INFLUXDB3_ADDR:-"http://0.0.0.0:8181"}
define_option server-name ${INFLUXDB3_SERVER_NAME:-"influx-srv"}

parse_args_and_validate "$@" || exit 1

token_path=$(get_option token-path)
db_name=$(get_option db-name)
influx_addr=$(get_option influx-addr)
server_name=$(get_option server-name)

# Use waitfile.sh to wait for the token file then execute the original entry point
/waitfile.sh \
   --file "$token_path" \
   --cmd 'token=$(cat "'"$token_path"'"); mkdir -p /app-root/config; cat > /app-root/config/config.json <<EOF
      {
         "DEFAULT_INFLUX_SERVER": "'"$influx_addr"'",
         "DEFAULT_INFLUX_DATABASE": "'"$db_name"'",
         "DEFAULT_API_TOKEN": "'"$token"'",
         "DEFAULT_SERVER_NAME": "'"$server_name"'"
      }
      EOF' \
   --exec "/app-root/entrypoint.sh --mode=admin"