SHELL := /bin/bash
# - Commands fail if any part of a pipeline fails (-o pipefail).
# - Commands are executed in a single shell (-c).
.SHELLFLAGS = -o pipefail -c

gateway_api_container_name := sensorflow-gateway-api
ingestion_api_container_name := sensorflow-ingestion-api
cli_container_name := sensorflow-cli
signer_container_name := sensorflow-signer
de_cli := docker exec -it $(cli_container_name)


#-----------------------------------------------------------------#
#                              Core                               #
#-----------------------------------------------------------------#

.PHONY: start stop reset

start:
	@docker compose -f ./ops/docker-compose.dev.yml up -d 

stop:
	@docker compose -f ./ops/docker-compose.dev.yml stop

reset:
	@docker compose -f ./ops/docker-compose.dev.yml down --volumes --remove-orphans


#-----------------------------------------------------------------#
#                             Reload                              #
#-----------------------------------------------------------------#

.PHONY: reload/gateway reload/ingestion reload/cli reload/signer

reload/gateway:
	$(call send_usr1_signal,$(gateway_api_container_name),run.dev.sh)

reload/ingestion:
	$(call send_usr1_signal,$(ingestion_api_container_name),run.dev.sh)

reload/cli:
	$(call send_usr1_signal,$(cli_container_name),run.dev.sh)

reload/signer:
	$(call send_usr1_signal,$(signer_container_name),run.dev.sh)

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

.PHONY: migrate/pg/new migrate/pg/up migrate/pg/down migrate/pg/goto \
	migrate/clickhouse/new migrate/clickhouse/up migrate/clickhouse/down migrate/clickhouse/goto

pg_migrate_path := ./internal/adapters/postgres/migrations

migrate/pg/new:
	@$(call de_migrate_create,${pg_migrate_path},${name})

migrate/pg/up:
	@$(call de_migrate_exec,${pg_migrate_path},PG_DATABASE_URL,up,${step})

migrate/pg/down:
	@$(call de_migrate_exec,${pg_migrate_path},PG_DATABASE_URL,down,${step})

migrate/pg/goto:
	@$(call de_migrate_exec,${pg_migrate_path},PG_DATABASE_URL,goto,${version})

clickhouse_migrate_path := ./internal/adapters/clickhouse/migrations

migrate/clickhouse/new:
	@$(call de_migrate_create,${clickhouse_migrate_path},${name})

migrate/clickhouse/up:
	@$(call de_migrate_exec,${clickhouse_migrate_path},CLICKHOUSE_DATABASE_URL,up,${step})

migrate/clickhouse/down:
	@$(call de_migrate_exec,${clickhouse_migrate_path},CLICKHOUSE_DATABASE_URL,down,${step})

migrate/clickhouse/goto:
	@$(call de_migrate_exec,${clickhouse_migrate_path},CLICKHOUSE_DATABASE_URL,goto,${version})

define de_migrate_create
	@$(de_cli) migrate create -seq -ext=.sql -dir=$(1) $(2)
endef

define de_migrate_exec
	@$(de_cli) sh -c 'migrate -path=$(1) -database "$$$(2)" $(3) $(4)'
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

