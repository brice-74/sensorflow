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
	SensorGatewayID      ulid.ULID `db:"sensor_gateway_id"`
	SensorGateway        *SensorGateway
	MeasurementProfileID ulid.ULID `db:"measurement_table_id"`
	MeasurementProfile   *MeasurementProfile
	Status               SensorStatus `db:"status"`
	Firmware             *string      `db:"firmware"`
	common.ULID
	common.Timestamps
	common.SoftDelete
	common.VersionUnix
}

// Gateway represents a central program or device that collects data from one or multiple sensors
// and forwards it to the system. It acts as the main entry point for sensor data ingestion.
type SensorGateway struct {
	TenantID        ulid.ULID `db:"tenant_id"`
	Tenant          *Tenant
	Name            string  `db:"name"`
	Location        *string `db:"location"`
	Firmware        *string `db:"firmware"`
	SensorInstances []*SensorInstance
	common.ULID
	common.Timestamps
	common.SoftDelete
	common.VersionUnix
}
