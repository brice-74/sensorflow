package domain

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type User struct {
	ID       ulid.ULID `json:"id" db:"id"`
	Email    string    `json:"email" db:"email"`
	Password *string   `json:"password" db:"password"`
	TenantID ulid.ULID `json:"tenant_id" db:"tenant_id"`
	common.Timestamps
	common.SoftDelete
}
