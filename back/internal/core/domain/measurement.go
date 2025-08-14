package domain

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type InfluxDataType uint8

const (
	InfluxDataTypeFloat = iota
	InfluxDataTypeInteger
	InfluxDataTypeBoolean
	InfluxDataTypeString
	InfluxDataTypeTimestamp
)

func (t InfluxDataType) String() string {
	switch t {
	case InfluxDataTypeFloat:
		return "float"
	case InfluxDataTypeInteger:
		return "int"
	case InfluxDataTypeBoolean:
		return "bool"
	case InfluxDataTypeString:
		return "string"
	case InfluxDataTypeTimestamp:
		return "timestamp"
	default:
		return "unknown"
	}
}

// MeasurementProfile represents a template for a type of measurement, defining its name, storage table.
type MeasurementProfile struct {
	ID          ulid.ULID `json:"id"`
	Name        string    `json:"name"`
	StorageName string    `json:"storage_name"`
	Description *string   `json:"description"`
	// Shared indicates whether this MeasurementProfile is shared across tenants.
	// If true, the corresponding storage table is managed internally and cannot be modified by tenants.
	Shared bool `json:"shared"`
	common.Timestamps
	common.SoftDelete
}

// MeasurementField represents a specific field (data point) within a MeasurementProfile.
type MeasurementField struct {
	ID          ulid.ULID      `json:"id"`
	ProfileID   ulid.ULID      `json:"profile_id"`
	Name        string         `json:"name"`
	StorageName string         `json:"storage_name"`
	Description *string        `json:"description"`
	Type        InfluxDataType `json:"type"`
	common.Timestamps
	common.SoftDelete
}

// MeasurementTag represents a tag (metadata) associated with a MeasurementProfile,
// used to categorize or filter measurement data.
type MeasurementTag struct {
	ID          ulid.ULID `json:"id"`
	ProfileID   ulid.ULID `json:"profile_id"`
	Name        string    `json:"name"`
	StorageName string    `json:"storage_name"`
	Description *string   `json:"description"`
	common.Timestamps
	common.SoftDelete
}
