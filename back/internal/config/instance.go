package config

import "github.com/brice-74/sensorflow/pkg/config"

type Instance struct {
	Identifier string
}

func (c *Instance) Define(loader *config.Loader) {
	loader.String(&c.Identifier, "api_identifier", "API_IDENTIFIER", "", "unique identifier for this backend instance").
		Required()
}
