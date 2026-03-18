package conversion

type TemperatureUnit int

const (
	CELSIUS TemperatureUnit = iota
	FAHRENHEIT
	KELVIN
)

func CelsiusToFahrenheit(c float64) float64 { return c*9/5 + 32 }
func FahrenheitToCelsius(f float64) float64 { return (f - 32) * 5 / 9 }
func KelvinToCelsius(k float64) float64     { return k - 273.15 }
func CelsiusToKelvin(c float64) float64     { return c + 273.15 }
func FahrenheitToKelvin(f float64) float64  { return (f-32)*5/9 + 273.15 }
func KelvinToFahrenheit(k float64) float64  { return (k-273.15)*9/5 + 32 }

func TemperatureToCelsius(value float64, unit TemperatureUnit) float64 {
	switch unit {
	case CELSIUS:
		return value
	case FAHRENHEIT:
		return FahrenheitToCelsius(value)
	case KELVIN:
		return KelvinToCelsius(value)
	default:
		return value
	}
}

func TemperatureToFahrenheit(value float64, unit TemperatureUnit) float64 {
	switch unit {
	case CELSIUS:
		return CelsiusToFahrenheit(value)
	case FAHRENHEIT:
		return value
	case KELVIN:
		return KelvinToFahrenheit(value)
	default:
		return value
	}
}

func TemperatureToKelvin(value float64, unit TemperatureUnit) float64 {
	switch unit {
	case CELSIUS:
		return CelsiusToKelvin(value)
	case FAHRENHEIT:
		return FahrenheitToKelvin(value)
	case KELVIN:
		return value
	default:
		return value
	}
}
