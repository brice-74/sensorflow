package redis

import "fmt"

type EntityKey string

func FormatEntityKey(entity EntityKey, identifier string) string {
	return fmt.Sprintf("%s:%s", entity, identifier)
}

func FormatHasManyKey(parent EntityKey, parentID string, child EntityKey) string {
	return fmt.Sprintf("%s:%s:%s", parent, parentID, child)
}
