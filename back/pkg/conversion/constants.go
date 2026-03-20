package conversion

import "math"

const (
	// Temperature
	KelvinOffset float64 = 273.15

	// Acceleration
	GToMS2Factor float64 = 9.80665

	// Gyro
	RadToDegFactor float64 = 180 / math.Pi
	DegToRadFactor float64 = math.Pi / 180
)
