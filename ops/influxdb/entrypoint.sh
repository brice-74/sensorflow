#!/bin/bash

exec influxdb3 serve \
  --object-store "${INFLUXDB_OBJECT_STORE}" \
  --node-id "node1" \
  --log-filter "${INFLUXDB_LOG_LEVEL:-info}" \
  --plugin-dir "${INFLUXDB_PLUGIN_DIR}" \
  --data-dir "${INFLUXDB_OBJECT_STORE_PATH}"