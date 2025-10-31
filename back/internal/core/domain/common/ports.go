package common

import (
	"fmt"
	"time"
)

type Identifiable[ID fmt.Stringer] interface {
	StringID() string
	GetID() ID
	SetID(ID)
	TouchID()
}

type Timestamped interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

type SoftDeleted interface {
	GetDeletedAt() *time.Time
	SetDeletedAt(*time.Time)
}

type Versioned[T any] interface {
	GetVersion() T
	SetVersion(T)
	TouchVersion()
}
