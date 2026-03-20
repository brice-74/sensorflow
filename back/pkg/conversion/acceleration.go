package conversion

type AccelUnit int

const (
	M_S2 AccelUnit = iota
	G
)

func MS2ToG(ms2 float64) float64 { return ms2 / GToMS2Factor }
func GToMS2(g float64) float64   { return g * GToMS2Factor }

func AccelToMS2(value float64, unit AccelUnit) float64 {
	switch unit {
	case M_S2:
		return value
	case G:
		return GToMS2(value)
	default:
		return value
	}
}

func AccelToG(value float64, unit AccelUnit) float64 {
	switch unit {
	case M_S2:
		return MS2ToG(value)
	case G:
		return value
	default:
		return value
	}
}
