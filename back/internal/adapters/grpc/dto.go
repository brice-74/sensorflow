package grpc

import (
	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

func AccelGyroDTO(sensorID uuid.UUID, tenantID uuid.UUID, m []*proto.AccelGyroMeasurement) []*domain.AccelGyroMeasurement {
	result := make([]*domain.AccelGyroMeasurement, len(m))
	for i, v := range m {
		result[i] = &domain.AccelGyroMeasurement{
			TenantID:      tenantID,
			SensorID:      sensorID,
			MeasureTime:   v.MeasuredAt.AsTime(),
			AccelX:        v.AccelX,
			AccelY:        v.AccelY,
			AccelZ:        v.AccelZ,
			GyroX:         v.GyroX,
			GyroY:         v.GyroY,
			GyroZ:         v.GyroZ,
			Temperature:   v.Temperature,
			VibrationRMS:  v.VibrationRms,
			VibrationPeak: v.VibrationPeak,
		}
	}
	return result
}
