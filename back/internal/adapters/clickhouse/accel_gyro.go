package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/google/uuid"
)

type (
	AccelGyroStdDynamicRepo        dynamicRepo
	AccelGyroIndustrialDynamicRepo dynamicRepo
	AccelGyroRealtimeDynamicRepo   dynamicRepo
)

func (r *AccelGyroStdDynamicRepo) NewBatch(ctx context.Context, dbID uuid.UUID) (Batch[domain.AccelGyroMeasurement], error) {
	return doWithRegistry(ctx, r.registry, dbID, func(ctx context.Context, conn clickhouse.Conn) (Batch[domain.AccelGyroMeasurement], error) {
		return accelGyroNewBatch(ctx, conn, "_std")
	})
}

func (r *AccelGyroIndustrialDynamicRepo) NewBatch(ctx context.Context, dbID uuid.UUID) (Batch[domain.AccelGyroMeasurement], error) {
	return doWithRegistry(ctx, r.registry, dbID, func(ctx context.Context, conn clickhouse.Conn) (Batch[domain.AccelGyroMeasurement], error) {
		return accelGyroNewBatch(ctx, conn, "_industrial")
	})
}

func (r *AccelGyroRealtimeDynamicRepo) NewBatch(ctx context.Context, dbID uuid.UUID) (Batch[domain.AccelGyroMeasurement], error) {
	return doWithRegistry(ctx, r.registry, dbID, func(ctx context.Context, conn clickhouse.Conn) (Batch[domain.AccelGyroMeasurement], error) {
		return accelGyroNewBatch(ctx, conn, "_realtime")
	})
}

type (
	AccelGyroStdRepo        repo
	AccelGyroIndustrialRepo repo
	AccelGyroRealtimeRepo   repo
)

func (r *AccelGyroStdRepo) NewBatch(ctx context.Context) (Batch[domain.AccelGyroMeasurement], error) {
	return accelGyroNewBatch(ctx, r.client.Client, "_std")
}

func (r *AccelGyroIndustrialRepo) NewBatch(ctx context.Context) (Batch[domain.AccelGyroMeasurement], error) {
	return accelGyroNewBatch(ctx, r.client.Client, "_industrial")
}

func (r *AccelGyroRealtimeRepo) NewBatch(ctx context.Context) (Batch[domain.AccelGyroMeasurement], error) {
	return accelGyroNewBatch(ctx, r.client.Client, "_realtime")
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

	return &AccelGyroBatch{batch}, nil
}

type AccelGyroBatch struct{ driver.Batch }

var _ Batch[domain.AccelGyroMeasurement] = (*AccelGyroBatch)(nil)

func (b *AccelGyroBatch) Append(rows ...domain.AccelGyroMeasurement) error {
	for _, m := range rows {
		err := b.Batch.Append(
			m.TenantID,
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
