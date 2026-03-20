package grpc

import (
	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/conversion"
)

func NormalizeAccelGyro(opts *proto.IngestAccelGyroOptions, measurements []*domain.AccelGyroMeasurement) {
	if *opts.AccelUnit == proto.AccelUnit_M_S2 &&
		*opts.GyroUnit == proto.GyroUnit_DEG_S &&
		*opts.TemperatureUnit == proto.TemperatureUnit_CELSIUS {
		return
	}

	doAccel := *opts.AccelUnit == proto.AccelUnit_G
	doGyro := *opts.GyroUnit == proto.GyroUnit_RAD_S
	doTempK := *opts.TemperatureUnit == proto.TemperatureUnit_KELVIN
	doTempF := *opts.TemperatureUnit == proto.TemperatureUnit_FAHRENHEIT

	for _, m := range measurements {
		if doAccel {
			m.AccelX *= conversion.GToMS2Factor
			m.AccelY *= conversion.GToMS2Factor
			m.AccelZ *= conversion.GToMS2Factor
		}

		if doGyro {
			m.GyroX *= conversion.RadToDegFactor
			m.GyroY *= conversion.RadToDegFactor
			m.GyroZ *= conversion.RadToDegFactor
		}

		if m.Temperature != nil {
			if doTempK {
				*m.Temperature -= conversion.KelvinOffset
			} else if doTempF {
				*m.Temperature = conversion.FahrenheitToCelsius(*m.Temperature)
			}
		}
	}
}
