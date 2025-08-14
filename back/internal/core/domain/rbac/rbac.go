package rbac

import (
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type Model struct {
	ID   int64  `json:"id"`
	Name string `json:"name"` // unique
}

// create unique index on every fields
type ModelHasRole struct {
	RoleID       ulid.ULID `json:"role_id"`
	ModelID      int64     `json:"model_id"`
	ModelInnerID ulid.ULID `json:"model_inner_id"`
}

// create unique index on (TenantID, GuardID, Name)
type Role struct {
	ID ulid.ULID `json:"id"`
	// An undefined tenantID means that the role is native and therefore cannot be modified externally.
	TenantID    *ulid.ULID `json:"tenant_id"`
	GuardID     ulid.ULID  `json:"guard_type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	common.Timestamps
}

// create unique index on (RoleID, PermissionID)
type RoleHasPermission struct {
	RoleID       ulid.ULID `json:"role_id"`
	PermissionID int64     `json:"permission_id"`
}

// create unique index on (ResourceActionID, ScopeID, GuardID)
type Permission struct {
	ID               int64  `json:"id"`
	ResourceActionID int64  `json:"resource_action_id"`
	ScopeID          *int64 `json:"scope_id"`
	GuardID          int64  `json:"guard_id"`
}

// Catalog of resources and possible actions also use as a reference for permissions
type ResourceAction struct {
	ID          int64  `json:"id"`
	ResourceID  int64  `json:"resource_id"`
	ActionID    int64  `json:"action_id"`
	Description string `json:"description"`
}

// Catalog of authorized scopes for each resource/action
type ResourceActionScope struct {
	ResourceActionID int64 `json:"resource_action_id"`
	ScopeID          int64 `json:"scope_id"`
}

// Catalog of authorized guards for each resource/action
type ResourceActionGuard struct {
	ResourceActionID int64 `json:"resource_action_id"`
	GuardID          int64 `json:"guard_id"`
}
