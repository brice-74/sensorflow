package cache

import (
	"strings"
)

type EntityKey string

// result -> entity:identifier
func FormatKey(entity EntityKey, identifier string) string {
	var b strings.Builder
	b.Grow(len(entity) + len(identifier) + 1)
	b.WriteString(string(entity))
	b.WriteByte(':')
	b.WriteString(identifier)
	return b.String()
}

// result -> parent:parent_identifier:child
func FormatHasManyKey(parent EntityKey, parentID string, child EntityKey) string {
	var b strings.Builder
	b.Grow(len(parent) + len(parentID) + len(child) + 2)
	b.WriteString(string(parent))
	b.WriteByte(':')
	b.WriteString(parentID)
	b.WriteByte(':')
	b.WriteString(string(child))
	return b.String()
}

const (
	SensorGatewayKey  EntityKey = "sensor_gateway"
	SensorInstanceKey EntityKey = "sensor_instance"
)
