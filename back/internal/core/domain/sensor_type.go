package domain

import "fmt"

type SensorTypeCode uint8

const (
	SensorTypeUnknown SensorTypeCode = iota
	SensorTypeAcc
	SensorTypeGyro
	SensorTypeTemp
)

func (t SensorTypeCode) IsValid() bool {
	switch t {
	case SensorTypeUnknown, SensorTypeAcc, SensorTypeGyro, SensorTypeTemp:
		return true
	}
	return false
}

func (t SensorTypeCode) String() string {
	switch t {
	case SensorTypeUnknown:
		return "Unknown"
	case SensorTypeAcc:
		return "Acc"
	case SensorTypeGyro:
		return "Gyro"
	case SensorTypeTemp:
		return "Temp"
	default:
		return fmt.Sprintf("SensorTypeCode(%d)", t)
	}
}

type SensorType struct {
	ID   int            `json:"id"`
	Code SensorTypeCode `json:"code"`
	Name string         `json:"name"`
}
