package domain

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type User struct {
	ID       ulid.ULID `db:"id"`
	Email    string    `db:"email"`
	Password *string   `db:"password"`
	TenantID ulid.ULID `db:"tenant_id"`
	common.Timestamps
	common.SoftDelete
}
