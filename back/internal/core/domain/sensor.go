package domain

import (
	"fmt"

	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorStatus uint8

const (
	SensorStatusUnknown SensorStatus = iota
	SensorStatusActive
	SensorStatusInactive
	SensorStatusError
)

func (s SensorStatus) IsValid() bool {
	switch s {
	case SensorStatusUnknown, SensorStatusActive, SensorStatusInactive, SensorStatusError:
		return true
	}
	return false
}

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
	ID                   ulid.ULID      `json:"id"`
	SensorGatewayID      ulid.ULID      `json:"sensor_gateway_id"`
	MeasurementProfileID ulid.ULID      `json:"measurement_profile_id"`
	Config               map[string]any `json:"config"`
	Status               SensorStatus   `json:"status"`
	Timestamps
	SoftDelete
}

// Gateway represents a central program or device that collects data from one or multiple sensors
// and forwards it to the system. It acts as the main entry point for sensor data ingestion.
type SensorGateway struct {
	ID       ulid.ULID      `json:"id"`
	TenantID ulid.ULID      `json:"tenant_id"`
	Name     string         `json:"name"`
	Location string         `json:"location"`
	Firmware string         `json:"firmware"`
	Metadata map[string]any `json:"metadata"`
	Timestamps
	SoftDelete
}
