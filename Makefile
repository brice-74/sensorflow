SHELL := /bin/bash
# - Commands fail if any part of a pipeline fails (-o pipefail).
# - Commands are executed in a single shell (-c).
.SHELLFLAGS = -o pipefail -c

gateway_api_container_name := sensorflow-gateway-api
ingestion_api_container_name := sensorflow-ingestion-api

.PHONY: dev/start dev/start/core dev/stop dev/reset dev/reload/gateway

dev/start:
	@docker compose -f ./ops/docker-compose.dev.yml up -d 

dev/start/core:
	@docker compose -f ./ops/docker-compose.dev.yml up -d gateway-api kafka influxdb

dev/stop:
	@docker compose -f ./ops/docker-compose.dev.yml stop

dev/reset:
	@docker compose -f ./ops/docker-compose.dev.yml down --volumes --remove-orphans

dev/reload/gateway:
	$(call reload_process,$(gateway_api_container_name),run.dev.sh)

dev/reload/ingestion:
	$(call reload_process,$(ingestion_api_container_name),run.dev.sh)

# generic function to send USR1 signal to containerize process
# parameters: 
# (1) container name
# (2) name of process target
define reload_process
	@docker exec -it $(1) bash -c 'kill -s USR1 "$$(pgrep -f $(2) | head -n 1)"'
endef

protoc/sensor:
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		back/internal/adapters/grpc/proto/sensor.proto