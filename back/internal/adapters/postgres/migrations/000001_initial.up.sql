CREATE TYPE sensor_plan_type AS ENUM (
   'accel_gyro_std',
   'accel_gyro_industrial',
   'accel_gyro_realtime'
);

CREATE TABLE tenant (
   id UUID PRIMARY KEY,
   name TEXT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE "user" (
   id UUID PRIMARY KEY,
   email TEXT NOT NULL UNIQUE,
   password TEXT,
   tenant_id UUID NOT NULL REFERENCES tenant(id),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE sensor_gateway (
   id UUID PRIMARY KEY,
   tenant_id UUID NOT NULL REFERENCES tenant(id),
   name TEXT NOT NULL,
   location TEXT,
   firmware TEXT,
   version BIGINT NOT NULL DEFAULT 0,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE sensor_instance (
   id UUID PRIMARY KEY,
   sensor_gateway_id UUID NOT NULL REFERENCES sensor_gateway(id),
   status SMALLINT NOT NULL DEFAULT 0 CHECK (status BETWEEN 0 AND 3),
   firmware TEXT,
   version BIGINT NOT NULL DEFAULT 0,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE measurement_logical_profile (
   id UUID PRIMARY KEY,
   tenant_id UUID NOT NULL REFERENCES tenant(id),
   name TEXT NOT NULL,
   version INTEGER NOT NULL CHECK (version BETWEEN 0 AND 65535),
   previous_version_id UUID REFERENCES measurement_logical_profile(id),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE measurement_field (
   id UUID PRIMARY KEY,
   logical_profile_id UUID NOT NULL REFERENCES measurement_logical_profile(id),
   name TEXT NOT NULL,
   type SMALLINT NOT NULL CHECK (type BETWEEN 0 AND 17),
   is_tag BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE measurement_database (
   id UUID PRIMARY KEY,
   tenant_id UUID REFERENCES tenant(id),
   name TEXT NOT NULL
);

CREATE TABLE measurement_table (
   id UUID PRIMARY KEY,
   database_id UUID NOT NULL REFERENCES measurement_database(id),
   name TEXT NOT NULL,
   hot_ttl TEXT,
   cold_ttl TEXT,
   partitioning_strategy TEXT NOT NULL,
   cold_volume_name TEXT
);

CREATE TABLE measurement_profile_binding (
   sensor_instance_id UUID NOT NULL REFERENCES sensor_instance(id),
   logical_profile_id UUID NOT NULL REFERENCES measurement_logical_profile(id),
   measurement_table_id UUID NOT NULL REFERENCES measurement_table(id),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ,
   PRIMARY KEY (sensor_instance_id, logical_profile_id, measurement_table_id)
);

CREATE TABLE sensor_plan (
   id UUID PRIMARY KEY,
   plan sensor_plan_type NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE sensor_plan_binding (
   sensor_instance_id UUID NOT NULL REFERENCES sensor_instance(id),
   sensor_plan_id UUID NOT NULL REFERENCES sensor_plan(id),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ,
   PRIMARY KEY (sensor_instance_id, sensor_plan_id)
);

CREATE INDEX idx_user_tenant_id ON "user"(tenant_id);
CREATE INDEX idx_sensor_gateway_tenant_id ON sensor_gateway(tenant_id);
CREATE INDEX idx_sensor_instance_sensor_gateway_id ON sensor_instance(sensor_gateway_id);
CREATE INDEX idx_measurement_logical_profile_tenant_id ON measurement_logical_profile(tenant_id);
CREATE INDEX idx_measurement_field_logical_profile_id ON measurement_field(logical_profile_id);
CREATE INDEX idx_measurement_database_tenant_id ON measurement_database(tenant_id);
CREATE INDEX idx_measurement_table_database_id ON measurement_table(database_id);
CREATE INDEX idx_measurement_profile_binding_sensor_instance_id ON measurement_profile_binding(sensor_instance_id);
CREATE INDEX idx_measurement_profile_binding_logical_profile_id ON measurement_profile_binding(logical_profile_id);
CREATE INDEX idx_measurement_profile_binding_measurement_table_id ON measurement_profile_binding(measurement_table_id);
CREATE INDEX idx_sensor_plan_binding_sensor_plan_id ON sensor_plan_binding(sensor_plan_id);

CREATE UNIQUE INDEX uq_sensor_plan_binding_active_sensor_instance
   ON sensor_plan_binding(sensor_instance_id)
   WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_measurement_profile_binding_active_sensor_instance
   ON measurement_profile_binding(sensor_instance_id)
   WHERE deleted_at IS NULL;