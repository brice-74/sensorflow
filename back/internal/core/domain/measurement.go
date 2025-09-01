package domain

import (
	"fmt"

	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type InfluxDataType uint8

const (
	InfluxDataTypeString = iota
	InfluxDataTypeFloat64
	InfluxDataTypeInt64
	InfluxDataTypeUint64
	InfluxDataTypeBool
	InfluxDataTypeTimestamp
)

func (t InfluxDataType) String() string {
	switch t {
	case InfluxDataTypeString:
		return "string"
	case InfluxDataTypeFloat64:
		return "float64"
	case InfluxDataTypeInt64:
		return "int64"
	case InfluxDataTypeUint64:
		return "uint64"
	case InfluxDataTypeBool:
		return "boolean"
	case InfluxDataTypeTimestamp:
		return "timestamp"
	default:
		return fmt.Sprintf("InfluxDataType(%d)", t)
	}
}

// MeasurementTable defines a specific storage table.
type MeasurementTable struct {
	ID          ulid.ULID `json:"id" db:"id"`
	ProfileID   ulid.ULID `json:"profile_id" db:"profile_id"`
	StorageName string    `json:"storage_name" db:"storage_name"`
	common.Timestamps
}

// MeasurementProfile represents a template for a type of measurement.
// An undefined tenant_id means that the profile is shared.
type MeasurementProfile struct {
	ID          ulid.ULID  `json:"id" db:"id"`
	TenantID    *ulid.ULID `json:"tenant_id" db:"tenant_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	common.Timestamps
	common.SoftDelete
}

// MeasurementField represents a specific field (data point) within a MeasurementProfile.
type MeasurementField struct {
	ID          ulid.ULID      `json:"id" db:"id"`
	ProfileID   ulid.ULID      `json:"profile_id" db:"profile_id"`
	Name        string         `json:"name" db:"name"`
	StorageName string         `json:"storage_name" db:"storage_name"`
	Description *string        `json:"description" db:"description"`
	Type        InfluxDataType `json:"type" db:"type"`
	common.Timestamps
	common.SoftDelete
}

// MeasurementTag represents a tag (metadata) associated with a MeasurementProfile,
// used to categorize or filter measurement data.
type MeasurementTag struct {
	ID          ulid.ULID `json:"id" db:"id"`
	ProfileID   ulid.ULID `json:"profile_id" db:"profile_id"`
	Name        string    `json:"name" db:"name"`
	StorageName string    `json:"storage_name" db:"storage_name"`
	Description *string   `json:"description" db:"description"`
	common.Timestamps
	common.SoftDelete
}
