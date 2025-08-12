package config

import (
	"fmt"
	"strconv"

	"github.com/brice-74/sensorflow/pkg/config"
)

type HTTP struct {
	Port string
}

func (c *HTTP) Define(loader *config.Loader) {
	loader.String(&c.Port, "api_port", "API_PORT", "", "exposed API port").
		Required().
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("port cannot be empty")
			}
			portNum, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("port must be a number")
			}
			if portNum < 1 || portNum > 65535 {
				return fmt.Errorf("port must be between 1 and 65535")
			}
			return nil
		})
}
