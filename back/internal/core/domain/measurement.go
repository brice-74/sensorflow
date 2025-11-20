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

type MeasurementProfileLogical struct {
	ID                ulid.ULID
	TenantID          *ulid.ULID // nil = shared
	Tenant            *Tenant
	PreviousVersionID *ulid.ULID
	PreviousVersion   *MeasurementProfileLogical
	Name              string
	Description       *string
	Version           uint16
	IsActive          bool
	Fields            []*MeasurementField
	common.Timestamps
	common.SoftDelete
}

type MeasurementProfilePhysical struct {
	ProfileID ulid.ULID
	DbType    string // "influx", "clickhouse"
	DBName    string
	TableName string // par version
	TTL       string // "7d", "30d"…
}

// MeasurementField represents a data point and can be an indexed tag for custom profiles.
type MeasurementField struct {
	ID          ulid.ULID
	ProfileID   ulid.ULID
	Profile     *MeasurementProfileLogical
	Name        string
	Description *string
	Type        DataType
	IsTag       bool
	common.Timestamps
	common.SoftDelete
}
