package domain

import "github.com/brice-74/sensorflow/pkg/ulid"

/* type GuardType uint8

const (
	GuardTypeUser GuardType = iota
	GuardTypeSensorGateway
	GuardTypeAPIGateway
) */

// faire index unique pour (TenantID, GuardID, Name)
type Role struct {
	ID ulid.ULID `json:"id"`
	// An undefined tenantID means that the role is native and therefore cannot be modified externally.
	TenantID    *ulid.ULID `json:"tenant_id"`
	GuardID     ulid.ULID  `json:"guard_type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Timestamps
}

type Guard struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Resource struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Action struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Scope struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// to see if it is useful and non-blocking in terms of mechanism
	//Level       *uint8 `json:"level"`
}

// faire index unique pour (RoleID, PermissionID)
type RoleHasPermission struct {
	RoleID       ulid.ULID `json:"role_id"`
	PermissionID int64     `json:"permission_id"` // correspond à une combinaison valide
}

// faire index unique pour (ResourceActionID, ScopeID, GuardID)
type Permission struct {
	ID               int64 `json:"id"`
	ResourceActionID int64 `json:"resource_action_id"`
	ScopeID          int64 `json:"scope_id"`
	GuardID          int64 `json:"guard_id"`
}

// Catalogue des ressources et actions possibles
// aussi utiliser en tant que référence pour les permission
type ResourceAction struct {
	ID         int64 `json:"id"`
	ResourceID int64 `json:"resource_id"`
	ActionID   int64 `json:"action_id"`
}

// Catalogue des scopes autorisés pour chaque resource/action
type ResourceActionScope struct {
	ResourceActionID int64 `json:"resource_action_id"`
	ScopeID          int64 `json:"scope_id"`
}

// Catalogue des guards autorisés pour chaque resource/action
type ResourceActionGuard struct {
	ResourceActionID int64 `json:"resource_action_id"`
	GuardID          int64 `json:"guard_id"`
}
