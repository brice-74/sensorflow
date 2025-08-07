SHELL := /bin/bash
# - Commands fail if any part of a pipeline fails (-o pipefail).
# - Commands are executed in a single shell (-c).
.SHELLFLAGS = -o pipefail -c

gateway_api_container_name := sensorflow-gateway-api

.PHONY: dev/up dev/down dev/reset dev/reload/gateway

dev/up:
	@docker compose -f ./ops/docker-compose.dev.yml up -d vault

dev/down:
	@docker compose -f ./ops/docker-compose.dev.yml down

dev/reset:
	@docker compose -f ./ops/docker-compose.dev.yml down --volumes --remove-orphans

dev/reload/gateway:
	$(call reload_process,$(gateway_api_container_name),run.dev.sh)

# generic function to send USR1 signal to containerize process
# parameters: 
# (1) container name
# (2) name of process target
define reload_process
	@docker exec -it $(1) bash -c 'kill -s USR1 "$$(pgrep -f $(2) | head -n 1)"'
endef