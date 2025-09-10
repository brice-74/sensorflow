package domain

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

// Tenant represents an organization or client using the platform, which can own multiple devices
type Tenant struct {
	ID   ulid.ULID `db:"id"`
	Name string    `db:"name"`
	common.Timestamps
	common.SoftDelete
}
