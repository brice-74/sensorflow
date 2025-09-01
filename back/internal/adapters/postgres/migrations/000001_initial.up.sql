CREATE TYPE sensor_status AS ENUM (
   'unknown',
   'active',
   'inactive',
   'error'
);

CREATE TYPE influx_data_type AS ENUM (
   'string',
   'float',
   'int64',
   'uint64',
   'bool',
   'timestamp'
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
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE measurement_profile (
   id UUID PRIMARY KEY,
   tenant_id UUID REFERENCES tenant(id),
   name TEXT NOT NULL,
   description TEXT,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE measurement_table (
   id UUID PRIMARY KEY,
   profile_id UUID NOT NULL REFERENCES measurement_profile(id),
   storage_name TEXT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE measurement_field (
   id UUID PRIMARY KEY,
   profile_id UUID NOT NULL REFERENCES measurement_profile(id),
   name TEXT NOT NULL,
   storage_name TEXT NOT NULL,
   description TEXT,
   type SMALLINT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE measurement_tag (
   id UUID PRIMARY KEY,
   profile_id UUID NOT NULL REFERENCES measurement_profile(id),
   name TEXT NOT NULL,
   storage_name TEXT NOT NULL,
   description TEXT,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE TABLE sensor_instance (
   id UUID PRIMARY KEY,
   sensor_gateway_id UUID NOT NULL REFERENCES sensor_gateway(id),
   measurement_table_id UUID NOT NULL REFERENCES measurement_table(id),
   status sensor_status NOT NULL DEFAULT 'unknown',
   firmware TEXT,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_user_tenant_id ON "user"(tenant_id);
CREATE INDEX idx_sensor_gateway_tenant_id ON sensor_gateway(tenant_id);
CREATE INDEX idx_sensor_instance_gateway_id ON sensor_instance(sensor_gateway_id);
CREATE INDEX idx_sensor_instance_table_id ON sensor_instance(measurement_table_id);