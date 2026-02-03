package domain

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Email    string    `db:"email"`
	Password *string   `db:"password"`
	TenantID uuid.UUID `db:"tenant_id"`
	common.Timestamps
	common.SoftDelete
}
