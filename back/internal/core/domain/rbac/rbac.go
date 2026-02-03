package rbac

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/google/uuid"
)

type Model struct {
	ID   int64
	Name string // unique
}

// create unique index on every fields
type ModelHasRole struct {
	RoleID       uuid.UUID
	ModelID      int64
	ModelInnerID uuid.UUID
}

// create unique index on (TenantID, GuardID, Name)
type Role struct {
	ID uuid.UUID
	// An undefined tenantID means that the role is native and therefore cannot be modified externally.
	TenantID    *uuid.UUID
	GuardID     uuid.UUID
	Name        string
	Description string
	common.Timestamps
}

// create unique index on (RoleID, PermissionID)
type RoleHasPermission struct {
	RoleID       uuid.UUID
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
