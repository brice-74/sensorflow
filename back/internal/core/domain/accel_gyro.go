package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccelGyroMeasurement struct {
	TenantID uuid.UUID `ch:"tenant_id"`
	SensorID uuid.UUID `ch:"sensor_id"`

	MeasureTime time.Time `ch:"measure_time"`
	IngestTime  time.Time `ch:"ingest_time"`

	// Accelerometer (m/s²)
	AccelX float64 `ch:"accel_x"`
	AccelY float64 `ch:"accel_y"`
	AccelZ float64 `ch:"accel_z"`

	// Gyroscope (rad/s)
	GyroX float64 `ch:"gyro_x"`
	GyroY float64 `ch:"gyro_y"`
	GyroZ float64 `ch:"gyro_z"`

	// Optional metrics
	Temperature   *float64 `ch:"temperature"`
	VibrationRMS  *float64 `ch:"vibration_rms"`
	VibrationPeak *float64 `ch:"vibration_peak"`
}
