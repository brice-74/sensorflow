SHELL := /bin/bash
# - Commands fail if any part of a pipeline fails (-o pipefail).
# - Commands are executed in a single shell (-c).
.SHELLFLAGS = -o pipefail -c

gateway_api_container_name := sensorflow-gateway-api
ingestion_api_container_name := sensorflow-ingestion-api
cli_container_name := sensorflow-cli
de_cli := docker exec -it $(cli_container_name)


#-----------------------------------------------------------------#
#                              Core                               #
#-----------------------------------------------------------------#

.PHONY: start start/core stop reset

start:
	@docker compose -f ./ops/docker-compose.dev.yml up -d 

start/core:
	@docker compose -f ./ops/docker-compose.dev.yml up -d cli gateway-api ingestion-api kafka influxdb postgres nginx

stop:
	@docker compose -f ./ops/docker-compose.dev.yml stop

reset:
	@docker compose -f ./ops/docker-compose.dev.yml down --volumes --remove-orphans


#-----------------------------------------------------------------#
#                             Reload                              #
#-----------------------------------------------------------------#

.PHONY: reload/gateway reload/ingestion reload/cli

reload/gateway:
	$(call send_usr1_signal,$(gateway_api_container_name),run.dev.sh)

reload/ingestion:
	$(call send_usr1_signal,$(ingestion_api_container_name),run.dev.sh)

reload/cli:
	$(call send_usr1_signal,$(cli_container_name),run.dev.sh)

# generic function to send USR1 signal to containerize process
# parameters: 
# (1) container name
# (2) name of process target
define send_usr1_signal
	@docker exec -it $(1) bash -c 'kill -s USR1 "$$(pgrep -f $(2) | head -n 1)"'
endef


#-----------------------------------------------------------------#
#                              Test                               #
#-----------------------------------------------------------------#

.PHONY: test/unit test/integration/postgres clean/testcache

test/unit: clean/testcache
	@$(call gotest,unit,$(func),$(path))

test/integration/postgres: clean/testcache
	@$(call gotest,integration_postgres,$(func),$(path))

define gotest
	@cd ./back && go test -p 1 -v -vet=off \
		-tags=$(1) \
		$(if $(2),-run $(2),) \
		$(or $(3),./...) \
		
endef

clean/testcache:
	@go clean -testcache


#-----------------------------------------------------------------#
#                             Migrate                             #
#-----------------------------------------------------------------#

.PHONY: migrate/new migrate/up migrate/down migrate/goto

migrate_path := ./internal/adapters/postgres/migrations

migrate/new:
	@$(de_cli) migrate create -seq -ext=.sql -dir=${migrate_path} ${name}

migrate/up:
	@$(call de_migrate,up,${step})

migrate/down:
	@$(call de_migrate,down,${step})

migrate/goto:
	@$(call de_migrate,goto,${version})

define de_migrate
	@$(de_cli) sh -c 'migrate -path=${migrate_path} -database "$$DATABASE_URL" $(1) $(2)'
endef


#-----------------------------------------------------------------#
#                             Other                               #
#-----------------------------------------------------------------#

.PHONY: protoc/sensor print/archi/back

protoc/sensor:
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		back/internal/adapters/grpc/proto/sensor.proto

print/archi/back:
	@tree -d -I 'logs' ./back

