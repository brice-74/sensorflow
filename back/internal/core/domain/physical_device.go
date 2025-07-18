package domain

import (
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type PhysicalDevice struct {
	ID       ulid.ULID      `json:"id"`
	TenantID ulid.ULID      `json:"tenant_id"`
	Name     string         `json:"name"`
	Location string         `json:"location"`
	Firmware string         `json:"firmware"`
	Metadata map[string]any `json:"metadata"`
	Timestamps
	SoftDelete
}
