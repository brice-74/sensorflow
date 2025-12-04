package domain

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type DataType uint8

const (
	DataTypeString DataType = iota
	DataTypeFloat64
	DataTypeInt64
	DataTypeUint64
	DataTypeBool
	DataTypeTimestamp
)

func (t DataType) GoString() string {
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
		return "bool"
	case DataTypeTimestamp:
		return "timestamp"
	default:
		return "unknown"
	}
}

type MeasurementLogicalProfile struct {
	common.ULID
	TenantID *ulid.ULID // nil = mutualisé
	Name     string
	Version  uint16
	IsActive bool

	PreviousVersionID *ulid.ULID
	Fields            []*MeasurementField

	common.Timestamps
	common.SoftDelete
}

type MeasurementField struct {
	common.ULID
	LogicalProfileID ulid.ULID
	Name             string
	Type             DataType
	IsTag            bool
}

type MeasurementDatabase struct {
	common.ULID
	Name     string
	Host     string
	User     string
	Password string
}

type MeasurementTable struct {
	common.ULID
	DatabaseID ulid.ULID

	Name                 string
	TTL                  string // "7d", "30d", etc.
	PartitioningStrategy string // "month", "week", "day"
	IsDedicated          bool   // table dédiée ou mutualisée
	EnableColdStorage    bool
}

type MeasurementProfileBinding struct {
	common.ULID
	SensorInstanceID      ulid.ULID
	LogicalProfileID      ulid.ULID
	MeasurementTableID    ulid.ULID
	MeasurementDatabaseID ulid.ULID

	EffectiveFrom int64  // timestamp de début
	EffectiveTo   *int64 // nil si actif
}
