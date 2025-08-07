#!/bin/bash
set -euo pipefail

influxdb3 serve \
   --object-store "${INFLUXDB3_OBJECT_STORE}" \
   --node-id "${INFLUXDB3_NODE_IDENTIFIER_PREFIX}" \
   --log-filter "${LOG_FILTER:-info}" \
   --plugin-dir "${INFLUXDB3_PLUGIN_DIR}" \
   --data-dir "${INFLUXDB3_DB_DIR}" \
   --http-bind "${INFLUXDB3_HTTP_BIND_ADDR}" &
influx_pid=$!

host="${INFLUXDB3_HOST:-localhost}"
port="${INFLUXDB3_PORT:-8181}"
addr="http://${INFLUXDB3_HTTP_BIND_ADDR}"

max_retries=10
count=0

# -----------------------------------------------------------------------------------------------
# Check service is up

echo "⏳  Waiting for InfluxDB 3 on ${host}:${port}..."

while ! timeout 1 bash -c "</dev/tcp/${host}/${port}" 2>/dev/null; do
   count=$((count + 1))
   if [ "$count" -ge "$max_retries" ]; then
      echo "❌  Port ${port} on ${host} not open after ${max_retries} attempts, exiting."
      exit 1
   fi
   echo "⏳  Port not open yet, retrying in 1 second... (${count}/${max_retries})"
   sleep 1
done

echo "✅  Service is up on ${host}:${port}"

# -----------------------------------------------------------------------------------------------
# Check/Create admin token

token_file="${INFLUXDB3_DB_DIR}/admin.token"

if [[ ! -f "$token_file" ]]; then
   echo "🔑  Creating admin token..."

   output=$(influxdb3 create token --admin --host "$addr" 2>&1) || {
      echo "❌  Failed to create token. Output:"
      echo "$output"
      exit 1
   }

   token=$(echo "$output" | grep -oP 'apiv3_\S+' | head -n1) || true

   if [[ -z "${token:-}" ]]; then
      echo "❌  Failed to extract token from output:"
      echo "$output"
      exit 1
   fi

   echo "$token" > "$token_file"
   echo "✅  Token created and saved to $token_file"
else
   token=$(cat "$token_file")
   echo "✅  Token already exists in $token_file"
fi

# -----------------------------------------------------------------------------------------------
# Check/Create database

echo "📦  Checking if database '${INFLUXDB3_DATABASE_NAME}' exists..."

if influxdb3 show databases \
   --host "$addr" \
   --token "$token" | grep -q "${INFLUXDB3_DATABASE_NAME}"; then
   echo "✅  Database '${INFLUXDB3_DATABASE_NAME}' already exists, skipping creation."
else
   echo "📦  Creating database '${INFLUXDB3_DATABASE_NAME}'..."
   output=$(influxdb3 create database "${INFLUXDB3_DATABASE_NAME}" \
      --host "$addr" \
      --token "$token" \
      --retention-period "${INFLUXDB3_RETENTION}" 2>&1) || {
         echo "❌  Failed to create database '${INFLUXDB3_DATABASE_NAME}'. Output:"
         echo "$output"
         exit 1
   }
   echo "✅  Database '${INFLUXDB3_DATABASE_NAME}' created successfully."
fi

wait "$influx_pid"