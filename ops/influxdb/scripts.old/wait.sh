#!/bin/bash
set -euo pipefail

source "${ARGPARSER_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/argparser.sh}"

define_option host ${INFLUXDB3_HOST:-"localhost"}
define_option port ${INFLUXDB3_PORT:-"8181"}
define_option max-retries ${INFLUXDB3_WAIT_MAX_RETRIES:-10}
define_option sleep-time ${INFLUXDB3_WAIT_SLEEP:-1}

parse_args "$@"

host=$(get_option host)
port=$(get_option port)
max_retries=$(get_option max-retries)
sleep_time=$(get_option sleep-time)
count=0

echo "⏳  Waiting for InfluxDB 3 on ${host}:${port}..."

while ! timeout 1 bash -c "</dev/tcp/${host}/${port}" 2>/dev/null; do
   count=$((count + 1))
   if [ "$count" -ge "$max_retries" ]; then
      echo "❌  Port ${port} on ${host} not open after ${max_retries} attempts, exiting."
      exit 1
   fi
   echo "⏳  Port not open yet, retrying in 1 second... (${count}/${max_retries})"
   sleep $sleep_time
done

echo "✅  Service is up on ${host}:${port}"