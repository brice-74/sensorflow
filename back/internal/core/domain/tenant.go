package domain

import (
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type Tenant struct {
	ID   ulid.ULID `json:"id"`
	Name string    `json:"name"`
	Timestamps
	SoftDelete
}
