package conversion

type GyroUnit int

const (
	DEG_S GyroUnit = iota
	RAD_S
)

func DegSToRadS(deg float64) float64 { return deg * DegToRadFactor }
func RadSToDegS(rad float64) float64 { return rad * RadToDegFactor }

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
