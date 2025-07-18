package domain

import (
	"time"
)

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ts *Timestamps) GetCreatedAt() time.Time {
	return ts.CreatedAt
}
func (ts *Timestamps) GetUpdatedAt() time.Time {
	return ts.UpdatedAt
}
func (ts *Timestamps) SetCreatedAt(t time.Time) {
	ts.CreatedAt = t
}
func (ts *Timestamps) SetUpdatedAt(t time.Time) {
	ts.UpdatedAt = t
}

type SoftDelete struct {
	DeletedAt *time.Time `json:"deleted_at"`
}

func (s *SoftDelete) GetDeletedAt() *time.Time {
	return s.DeletedAt
}
func (s *SoftDelete) SetDeletedAt(t *time.Time) {
	s.DeletedAt = t
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
