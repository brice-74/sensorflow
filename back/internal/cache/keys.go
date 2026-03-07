package cache

type (
	EntityKey  string
	LogicalKey string
)

const (
	SensorGatewayKey     EntityKey = "sensor_gateway"
	SensorPlanBindingKey EntityKey = "sensor_plan_binding"
	SensorInstanceKey    EntityKey = "sensor_instance"

	WithSensorInstancesKey LogicalKey = "with_sensor_instances"
)

// FormatKey returns the classic key of an entity
// Example: {entity}:{id}
func FormatKey(entity EntityKey, id string) string {
	return string(entity) + ":" + id
}

// FormatKeys returns the classic keys for a list of IDs.
func FormatKeys(entity EntityKey, ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = FormatKey(entity, id)
	}
	return keys
}

// FormatHasManyKey returns the key of a "has many" relationship.
// Example : {entity}:{id}:keys:{related_entity}
func FormatHasManyKey(parent EntityKey, parentID string, child EntityKey) string {
	return string(parent) + ":" + parentID + ":keys:" + string(child)
}

// FormatHydratedKey returns the key for a hydrated entity.
// Example : {entity}:{id}:hydrate:{logical_relational_naming}
func FormatHydratedKey(entity EntityKey, id string, logicalKey LogicalKey) string {
	return string(entity) + ":" + id + ":hydrate:" + string(logicalKey)
}

// FormatHydratedKeys returns multiple hydrated keys for a slice of IDs.
func FormatHydratedKeys(entity EntityKey, ids []string, logicalKey LogicalKey) []string {
	if len(ids) == 0 {
		return nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = FormatHydratedKey(entity, id, logicalKey)
	}
	return keys
}

// FormatActivePlanKey returns the Redis key for the active plan of a sensor instance
// Example : sensor_instance:{sensor_id}:active_plan
func FormatActivePlanKey(sensorID string) string {
	return string(SensorInstanceKey) + ":" + sensorID + ":active_plan"
}
