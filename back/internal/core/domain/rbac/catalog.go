package rbac

type GuardType uint8

const (
	GuardTypeUnknow = iota
	GuardTypeUser
	GuardTypeSensorGateway
	GuardTypeAPIGateway
)

type Guard struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"` // unique
	Description string    `json:"description"`
	Type        GuardType `json:"type"`
}

type ActionType uint8

const (
	ActionTypeUnknow = iota
	ActionTypeRead
	ActionTypeCreate
	ActionTypeUpdate
	ActionTypeSoftDelete
	ActionTypeHardDelete
	ActionTypeRestore
	ActionTypeExecute
)

type Action struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"` // unique
	Description string     `json:"description"`
	Type        ActionType `json:"type"`
}

type ScopeType uint8

type Scope struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"` // unique
	Description string    `json:"description"`
	Type        ScopeType `json:"type"`
	// to see if it is useful and non-blocking in terms of mechanism
	//Level       *uint8 `json:"level"`
}

type ResourceType uint8

type Resource struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"` // unique
	Description string       `json:"description"`
	Type        ResourceType `json:"type"`
}
