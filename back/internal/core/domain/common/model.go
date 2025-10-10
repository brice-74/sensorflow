package common

import "time"

type Timestamped interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

type Timestamps struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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

type SoftDeleted interface {
	GetDeletedAt() *time.Time
	SetDeletedAt(*time.Time)
}

type SoftDelete struct {
	DeletedAt *time.Time `db:"deleted_at"`
}

func (s *SoftDelete) GetDeletedAt() *time.Time {
	return s.DeletedAt
}
func (s *SoftDelete) SetDeletedAt(t *time.Time) {
	s.DeletedAt = t
}

type VersionedUnix interface {
	GetVersion() int64
	SetVersion(int64)
	Touch()
}

type VersionUnix struct {
	Version int64 `db:"version"`
}

func (v *VersionUnix) GetVersion() int64 {
	return v.Version
}

func (v *VersionUnix) SetVersion(ver int64) {
	v.Version = ver
}

func (v *VersionUnix) Touch() {
	v.Version = time.Now().UnixNano()
}
