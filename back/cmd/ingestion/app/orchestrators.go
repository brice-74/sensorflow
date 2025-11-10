package app

import (
	"github.com/brice-74/sensorflow/internal/orchestrators"
	"github.com/brice-74/sensorflow/internal/ports"
)

type Orchestrators struct {
	SensorGateway ports.SensorGatewayOrchestrator
}

func NewOrchestrators() *Orchestrators {
	orcts := Orchestrators{
		SensorGateway: &orchestrators.SensorGateway{},
	}
	return &orcts
}
