package domain

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type User struct {
	ID       ulid.ULID `json:"id"`
	Email    string    `json:"email"`
	Password *string   `json:"password"`
	TenantID ulid.ULID `json:"tenant_id"`
	common.Timestamps
	common.SoftDelete
}
