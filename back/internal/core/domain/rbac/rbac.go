package rbac

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type Model struct {
	ID   int64
	Name string // unique
}

// create unique index on every fields
type ModelHasRole struct {
	RoleID       ulid.ULID
	ModelID      int64
	ModelInnerID ulid.ULID
}

// create unique index on (TenantID, GuardID, Name)
type Role struct {
	ID ulid.ULID
	// An undefined tenantID means that the role is native and therefore cannot be modified externally.
	TenantID    *ulid.ULID
	GuardID     ulid.ULID
	Name        string
	Description string
	common.Timestamps
}

// create unique index on (RoleID, PermissionID)
type RoleHasPermission struct {
	RoleID       ulid.ULID
	PermissionID int64
}

// create unique index on (ResourceActionID, ScopeID, GuardID)
type Permission struct {
	ID               int64
	ResourceActionID int64
	ScopeID          *int64
	GuardID          int64
}

// Catalog of resources and possible actions also use as a reference for permissions
type ResourceAction struct {
	ID          int64
	ResourceID  int64
	ActionID    int64
	Description string
}

// Catalog of authorized scopes for each resource/action
type ResourceActionScope struct {
	ResourceActionID int64
	ScopeID          int64
}

// Catalog of authorized guards for each resource/action
type ResourceActionGuard struct {
	ResourceActionID int64
	GuardID          int64
}
