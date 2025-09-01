package domain

import (
	"fmt"

	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorStatus uint8

const (
	SensorStatusUnknown SensorStatus = iota
	SensorStatusActive
	SensorStatusInactive
	SensorStatusError
)

func (s SensorStatus) String() string {
	switch s {
	case SensorStatusUnknown:
		return "Unknown"
	case SensorStatusActive:
		return "Active"
	case SensorStatusInactive:
		return "Inactive"
	case SensorStatusError:
		return "Error"
	default:
		return fmt.Sprintf("SensorStatus(%d)", s)
	}
}

// SensorInstance represents an instance of a sensor attached to a device
type SensorInstance struct {
	ID                 ulid.ULID    `json:"id" db:"id"`
	SensorGatewayID    ulid.ULID    `json:"sensor_gateway_id" db:"sensor_gateway_id"`
	MeasurementTableID ulid.ULID    `json:"measurement_table_id" db:"measurement_table_id"`
	Status             SensorStatus `json:"status" db:"status"`
	Firmware           *string      `json:"firmware" db:"firmware"`
	common.Timestamps
	common.SoftDelete
}

// Gateway represents a central program or device that collects data from one or multiple sensors
// and forwards it to the system. It acts as the main entry point for sensor data ingestion.
type SensorGateway struct {
	ID       ulid.ULID `json:"id" db:"id"`
	TenantID ulid.ULID `json:"tenant_id" db:"tenant_id"`
	Name     string    `json:"name" db:"name"`
	Location *string   `json:"location" db:"location"`
	Firmware *string   `json:"firmware" db:"firmware"`
	common.Timestamps
	common.SoftDelete
}
