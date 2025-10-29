package common

import (
	"time"

	"github.com/brice-74/sensorflow/pkg/ulid"
)

type ULID struct {
	ID ulid.ULID `db:"id"`
}

var _ Identifiable[ulid.ULID] = (*ULID)(nil)

func (u *ULID) Identifier() ulid.ULID {
	return u.ID
}

func (u *ULID) SetIdentifier(id ulid.ULID) {
	u.ID = id
}

func (u *ULID) Touch() {
	u.ID = ulid.NewOrdered()
}

func (u *ULID) String() string {
	return u.ID.String()
}

func (u *ULID) StringIdentifier() string {
	return u.ID.String()
}

type Timestamps struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var _ Timestamped = (*Timestamps)(nil)

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
	DeletedAt *time.Time `db:"deleted_at"`
}

var _ SoftDeleted = (*SoftDelete)(nil)

func (s *SoftDelete) GetDeletedAt() *time.Time {
	return s.DeletedAt
}
func (s *SoftDelete) SetDeletedAt(t *time.Time) {
	s.DeletedAt = t
}

type VersionUnix struct {
	Version int64 `db:"version"`
}

var _ Versioned[int64] = (*VersionUnix)(nil)

func (v *VersionUnix) GetVersion() int64 {
	return v.Version
}

func (v *VersionUnix) SetVersion(ver int64) {
	v.Version = ver
}

func (v *VersionUnix) Touch() {
	v.Version = time.Now().UnixNano()
}
