#!/bin/bash
set -euo pipefail

influxdb3 serve \
   --object-store "${INFLUXDB3_OBJECT_STORE}" \
   --node-id "${INFLUXDB3_NODE_IDENTIFIER_PREFIX}" \
   --log-filter "${LOG_FILTER:-info}" \
   --plugin-dir "${INFLUXDB3_PLUGIN_DIR}" \
   --data-dir "${INFLUXDB3_DB_DIR}" \
   --http-bind "${INFLUXDB3_HTTP_BIND_ADDR}"