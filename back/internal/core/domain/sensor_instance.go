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

type SensorInstance struct {
	ID       ulid.ULID      `json:"id"`
	DeviceID ulid.ULID      `json:"device_id"`
	TypeID   ulid.ULID      `json:"type_id"`
	Config   map[string]any `json:"config"`
	Status   SensorStatus   `json:"status"`
	Timestamps
	SoftDelete
}
