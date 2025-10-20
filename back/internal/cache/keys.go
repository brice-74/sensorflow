package cache

type EntityKey string

const (
	SensorGatewayKey  EntityKey = "sensor_gateway"
	SensorInstanceKey EntityKey = "sensor_instance"
)

func FormatKey(entity EntityKey, identifier string) string {
	return string(entity) + ":" + identifier
}

func FormatKeys(entity EntityKey, ids []string) []string {
	n := len(ids)
	if n == 0 {
		return nil
	}

	keys := make([]string, n)
	for i, id := range ids {
		keys[i] = FormatKey(entity, id)
	}
	return keys
}

func FormatHasManyKey(parent EntityKey, parentID string, child EntityKey) string {
	return string(parent) + ":" + parentID + ":" + string(child)
}
