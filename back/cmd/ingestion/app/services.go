package app

import (
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/internal/core/service"
)

type Services struct {
	SensorGateway ports.SensorGatewayService
}

func NewServices(repos *Repositories, uows *UnitOfWorks) *Services {
	sensorGateway := service.NewSensorGateway(uows.ConnectionUnitOfWork, repos.SensorGateway, repos.SensorInstance)

	services := Services{
		SensorGateway: sensorGateway,
	}

	return &services
}
