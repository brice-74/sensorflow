CREATE TABLE IF NOT EXISTS sensorflow_generic.accel_gyro_realtime
(
  tenant_id UUID,
  device_id UUID,
  sensor_id UUID,

  measure_time DateTime64(6),
  ingest_time DateTime64(6) DEFAULT now64(6),

  accel_x Float64,
  accel_y Float64,
  accel_z Float64,

  gyro_x Float64,
  gyro_y Float64,
  gyro_z Float64,

  temperature Nullable(Float64),
  vibration_rms Nullable(Float64),
  vibration_peak Nullable(Float64)
)
ENGINE = MergeTree
PARTITION BY toYYYYMMDD(measure_time)
ORDER BY (tenant_id, device_id, sensor_id, measure_time)
TTL
  measure_time + INTERVAL 2 DAY DELETE
SETTINGS index_granularity = 16384;