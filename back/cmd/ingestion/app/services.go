package app

import (
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/internal/core/service"
)

type Services struct {
	SensorGateway ports.SensorGatewayService
}

func NewServices(repos *Repositories) *Services {
	sensorGateway := service.NewSensorGateway(repos.SensorGateway, repos.SensorInstance)

	services := Services{
		SensorGateway: sensorGateway,
	}

	return &services
}
