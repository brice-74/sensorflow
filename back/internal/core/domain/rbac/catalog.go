package rbac

type GuardType uint8

const (
	GuardTypeUnknow = iota
	GuardTypeUser
	GuardTypeSensorGateway
	GuardTypeAPIGateway
)

type Guard struct {
	ID          int64
	Name        string // unique
	Description string
	Type        GuardType
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
	ID          int64
	Name        string // unique
	Description string
	Type        ActionType
}

type ScopeType uint8

type Scope struct {
	ID          int64
	Name        string // unique
	Description string
	Type        ScopeType
	// to see if it is useful and non-blocking in terms of mechanism
	//Level       *uint8
}

type ResourceType uint8

type Resource struct {
	ID          int64
	Name        string // unique
	Description string
	Type        ResourceType
}
