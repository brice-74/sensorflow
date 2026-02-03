package domain

import (
	"fmt"

	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/google/uuid"
)

type DataType uint8

const (
	DataTypeFloat64 DataType = iota
	DataTypeInt64
	DataTypeUint64
	DataTypeFloat32
	DataTypeInt32
	DataTypeUint32

	DataTypeBool

	DataTypeTimestamp
	DataTypeDate
	DataTypeDateTime
	DataTypeDuration

	DataTypeString
	DataTypeLowCardinalityString
	DataTypeUUID

	DataTypeDecimal32
	DataTypeDecimal64
	DataTypeDecimal128
)

func (t DataType) GoString() string {
	switch t {
	case DataTypeFloat64:
		return "float64"
	case DataTypeInt64:
		return "int64"
	case DataTypeUint64:
		return "uint64"
	case DataTypeFloat32:
		return "float32"
	case DataTypeInt32:
		return "int32"
	case DataTypeUint32:
		return "uint32"
	case DataTypeBool:
		return "bool"
	case DataTypeTimestamp:
		return "timestamp"
	case DataTypeDate:
		return "date"
	case DataTypeDateTime:
		return "datetime"
	case DataTypeDuration:
		return "duration"
	case DataTypeString:
		return "string"
	case DataTypeLowCardinalityString:
		return "lowcardstring"
	case DataTypeUUID:
		return "uuid"
	case DataTypeDecimal32:
		return "decimal32"
	case DataTypeDecimal64:
		return "decimal64"
	case DataTypeDecimal128:
		return "decimal128"
	default:
		return "unknown"
	}
}

// Optional helper for ClickHouse type names
func (t DataType) ClickHouseType() string {
	switch t {
	case DataTypeFloat64:
		return "Float64"
	case DataTypeFloat32:
		return "Float32"
	case DataTypeInt64:
		return "Int64"
	case DataTypeInt32:
		return "Int32"
	case DataTypeUint64:
		return "UInt64"
	case DataTypeUint32:
		return "UInt32"
	case DataTypeBool:
		return "UInt8"
	case DataTypeTimestamp:
		return "DateTime64(3)"
	case DataTypeDate:
		return "Date"
	case DataTypeDateTime:
		return "DateTime"
	case DataTypeDuration:
		return "Int64"
	case DataTypeString:
		return "String"
	case DataTypeLowCardinalityString:
		return "LowCardinality(String)"
	case DataTypeUUID:
		return "UUID"
	case DataTypeDecimal32:
		return "Decimal32(9)"
	case DataTypeDecimal64:
		return "Decimal64(18)"
	case DataTypeDecimal128:
		return "Decimal128(38)"
	default:
		return "String"
	}
}

// Optional debug
func (t DataType) String() string {
	return fmt.Sprintf("%s (%d)", t.GoString(), t)
}

type MeasurementLogicalProfile struct {
	common.UUID
	TenantID uuid.UUID
	Name     string
	Version  uint16
	IsActive bool

	PreviousVersionID *uuid.UUID
	Fields            []*MeasurementField

	common.Timestamps
	common.SoftDelete
}

type MeasurementField struct {
	common.UUID
	LogicalProfileID uuid.UUID
	Name             string
	Type             DataType
	IsTag            bool
}

type MeasurementDatabase struct {
	common.UUID
	Name string
}

type MeasurementTable struct {
	common.UUID

	DatabaseID uuid.UUID
	Name       string

	Plan                 string  // ex: "custom", "shared_generic"
	HotTTL               *string // ex: "7d"
	ColdTTL              *string // ex: "90d"
	PartitioningStrategy string  // ex: "month", "week", "day"
	ColdVolumeName       *string
}

type MeasurementProfileBinding struct {
	SensorInstanceID      uuid.UUID
	LogicalProfileID      *uuid.UUID
	MeasurementTableID    uuid.UUID
	MeasurementDatabaseID uuid.UUID

	common.Timestamps
	common.SoftDelete
}
