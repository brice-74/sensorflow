package factory

import (
	"time"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type TenantSpec struct {
	ID        *uuid.UUID
	Name      *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type GatewaySpec struct {
	ID          *uuid.UUID
	Name        *string
	Location    *string
	Firmware    *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
	Certificate *CertificateSpec
}

type InstanceSpec struct {
	ID                      *uuid.UUID
	Status                  *domain.SensorStatus
	Firmware                *string
	CreatedAt               *time.Time
	UpdatedAt               *time.Time
	DeletedAt               *time.Time
	ActiveSensorPlanBinding *domain.SensorPlanBinding
}

type PlanSpec struct {
	ID        *uuid.UUID
	Plan      *domain.SensorPlanType
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type BindingSpec struct {
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type CertificateSpec struct {
	CommonName *string
	DNSNames   []string
	ValidFor   *time.Duration
}

type CertificateRequest struct {
	CommonName string
	DNSNames   []string
	ValidFor   time.Duration
}

type GatewayCertificate struct {
	GatewayID uuid.UUID
	Request   CertificateRequest
}
