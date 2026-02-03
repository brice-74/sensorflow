package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/errors"
)

type Repo[T any] interface {
	NewBatch(ctx context.Context) (Batch[T], error)
}

type Batch[T any] interface {
	Append(...T) error
	Flush() error
	Send() error
}

type repo struct {
	conn clickhouse.Conn
}

type (
	AccelGyroStdRepo        repo
	AccelGyroIndustrialRepo repo
	AccelGyroRealtimeRepo   repo
)

func (r *AccelGyroStdRepo) NewBatch(ctx context.Context) (Batch[domain.AccelGyroMeasurement], error) {
	return accelGyroNewBatch(ctx, r.conn, "_std")
}

func (r *AccelGyroIndustrialRepo) NewBatch(ctx context.Context) (Batch[domain.AccelGyroMeasurement], error) {
	return accelGyroNewBatch(ctx, r.conn, "_industrial")
}

func (r *AccelGyroRealtimeRepo) NewBatch(ctx context.Context) (Batch[domain.AccelGyroMeasurement], error) {
	return accelGyroNewBatch(ctx, r.conn, "_realtime")
}

func accelGyroNewBatch(
	ctx context.Context,
	conn clickhouse.Conn,
	suffix string,
) (Batch[domain.AccelGyroMeasurement], error) {
	if conn == nil {
		return nil, errors.WrapMsg("nil conn")
	}

	var tablename string = "accel_gyro" + suffix

	batch, err := conn.PrepareBatch(ctx, `
		INSERT INTO `+tablename+` (
			tenant_id,
			device_id,
			sensor_id,
			measure_time,
			ingest_time,
			accel_x,
			accel_y,
			accel_z,
			gyro_x,
			gyro_y,
			gyro_z,
			temperature,
			vibration_rms,
			vibration_peak
		)
	`)
	if err != nil {
		return nil, err
	}

	return &AccelGyrobatch{batch}, nil
}

type AccelGyrobatch struct{ driver.Batch }

func (b *AccelGyrobatch) Append(rows ...domain.AccelGyroMeasurement) error {
	for _, m := range rows {
		err := b.Batch.Append(
			m.TenantID,
			m.DeviceID,
			m.SensorID,
			m.MeasureTime,
			m.IngestTime,
			m.AccelX,
			m.AccelY,
			m.AccelZ,
			m.GyroX,
			m.GyroY,
			m.GyroZ,
			m.Temperature,
			m.VibrationRMS,
			m.VibrationPeak,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
