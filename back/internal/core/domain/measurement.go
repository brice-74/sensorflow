package domain

import (
	"fmt"

	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type DataType uint8

const (
	DataTypeString = iota
	DataTypeFloat64
	DataTypeInt64
	DataTypeUint64
	DataTypeBool
	DataTypeTimestamp
)

func (t DataType) String() string {
	switch t {
	case DataTypeString:
		return "string"
	case DataTypeFloat64:
		return "float64"
	case DataTypeInt64:
		return "int64"
	case DataTypeUint64:
		return "uint64"
	case DataTypeBool:
		return "boolean"
	case DataTypeTimestamp:
		return "timestamp"
	default:
		return fmt.Sprintf("DataType(%d)", t)
	}
}

// MeasurementProfile represents a template for a set of measurements.
// If TenantID is nil, the profile is shared across tenants.
type MeasurementProfile struct {
	ID          ulid.ULID
	TenantID    *ulid.ULID
	Tenant      *Tenant
	Name        string
	Description *string
	Groups      []*MeasurementGroup
	Fields      []*MeasurementField
	Tags        []*MeasurementTag
	common.Timestamps
	common.SoftDelete
}

// MeasurementGroup represents a logical set of measurements.
type MeasurementGroup struct {
	ID        ulid.ULID
	ProfileID ulid.ULID
	Profile   *MeasurementProfile
	Name      string
	common.Timestamps
}

// MeasurementField represents a data point.
type MeasurementField struct {
	ID          ulid.ULID
	ProfileID   ulid.ULID
	Profile     *MeasurementProfile
	Name        string
	Description *string
	Type        DataType
	common.Timestamps
	common.SoftDelete
}

// MeasurementTag represents a tag or indexed metadata.
type MeasurementTag struct {
	ID          ulid.ULID
	ProfileID   ulid.ULID
	Profile     *MeasurementProfile
	Name        string
	Description *string
	common.Timestamps
	common.SoftDelete
}
