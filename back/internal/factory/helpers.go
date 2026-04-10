package factory

import (
	"strings"
	"time"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

func (f *Factory) timeOrNow(value *time.Time) time.Time {
	if value != nil {
		return value.UTC()
	}
	return f.now()
}

func (f *Factory) uuidOrNew(value *uuid.UUID) uuid.UUID {
	if value != nil && *value != uuid.Nil {
		return *value
	}
	return uuid.New()
}

func stringOrDefault(value *string, fallback string) string {
	if value != nil && strings.TrimSpace(*value) != "" {
		return *value
	}
	return fallback
}

func durationOrDefault(value *time.Duration, fallback time.Duration) time.Duration {
	if value != nil && *value > 0 {
		return *value
	}
	return fallback
}

func sensorStatusOrDefault(value *domain.SensorStatus, fallback domain.SensorStatus) domain.SensorStatus {
	if value != nil {
		return *value
	}
	return fallback
}

func planTypeOrDefault(value *domain.SensorPlanType, fallback domain.SensorPlanType) domain.SensorPlanType {
	if value != nil {
		return *value
	}
	return fallback
}
