#!/bin/bash

chown -R clickhouse:clickhouse /var/lib/clickhouse/cold_data

exec /entrypoint.sh