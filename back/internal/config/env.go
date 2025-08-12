package config

import (
	"fmt"

	"github.com/brice-74/sensorflow/pkg/config"
)

const (
	EnvUnknow Env = iota
	EnvLocal
	EnvPreprod
	EnvProd
)

type Env uint8

func (e Env) String() string {
	switch e {
	case EnvLocal:
		return "local"
	case EnvPreprod:
		return "preprod"
	case EnvProd:
		return "prod"
	default:
		return "unknown"
	}
}

func EnvFromString(value string) (Env, error) {
	switch value {
	case "local":
		return EnvLocal, nil
	case "preprod":
		return EnvPreprod, nil
	case "prod":
		return EnvProd, nil
	default:
		return EnvUnknow, fmt.Errorf("unknow env: %s", value)
	}
}

func (e *Env) Define(loader *config.Loader) {
	config.AddOption(loader, e, "env", "ENV", "application environment (local|preprod|prod)", 0, EnvFromString).
		Required()
}
