package conversion

import "math"

type GyroUnit int

const (
	DEG_S GyroUnit = iota
	RAD_S
)

func DegSToRadS(deg float64) float64 { return deg * math.Pi / 180 }
func RadSToDegS(rad float64) float64 { return rad * 180 / math.Pi }

func GyroToDegS(value float64, unit GyroUnit) float64 {
	switch unit {
	case DEG_S:
		return value
	case RAD_S:
		return RadSToDegS(value)
	default:
		return value
	}
}

func GyroToRadS(value float64, unit GyroUnit) float64 {
	switch unit {
	case DEG_S:
		return DegSToRadS(value)
	case RAD_S:
		return value
	default:
		return value
	}
}
