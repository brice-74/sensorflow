package factory

import (
	"strings"
	"time"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

func (f *Factory) Tenant(spec TenantSpec) *domain.Tenant {
	id := f.uuidOrNew(spec.ID)
	tenant := &domain.Tenant{
		ID:   id,
		Name: stringOrDefault(spec.Name, defaultName("tenant", id)),
	}
	tenant.CreatedAt = f.timeOrNow(spec.CreatedAt)
	tenant.UpdatedAt = f.timeOrNow(spec.UpdatedAt)
	tenant.DeletedAt = spec.DeletedAt
	return tenant
}

func (f *Factory) Gateway(tenantID uuid.UUID, spec GatewaySpec) *domain.SensorGateway {
	id := f.uuidOrNew(spec.ID)
	gateway := &domain.SensorGateway{
		TenantID: tenantID,
		Name:     stringOrDefault(spec.Name, defaultName("gateway", id)),
		Location: spec.Location,
		Firmware: spec.Firmware,
	}
	gateway.ID = id
	gateway.CreatedAt = f.timeOrNow(spec.CreatedAt)
	gateway.UpdatedAt = f.timeOrNow(spec.UpdatedAt)
	gateway.DeletedAt = spec.DeletedAt
	return gateway
}

func (f *Factory) Instance(gatewayID uuid.UUID, spec InstanceSpec) *domain.SensorInstance {
	id := f.uuidOrNew(spec.ID)
	instance := &domain.SensorInstance{
		SensorGatewayID:         gatewayID,
		Status:                  sensorStatusOrDefault(spec.Status, domain.SensorStatusActive),
		Firmware:                spec.Firmware,
		ActiveSensorPlanBinding: spec.ActiveSensorPlanBinding,
	}
	instance.ID = id
	instance.CreatedAt = f.timeOrNow(spec.CreatedAt)
	instance.UpdatedAt = f.timeOrNow(spec.UpdatedAt)
	instance.DeletedAt = spec.DeletedAt
	instance.Version = f.now().UnixNano()
	return instance
}

func (f *Factory) Plan(spec PlanSpec) *domain.SensorPlan {
	id := f.uuidOrNew(spec.ID)
	pl := &domain.SensorPlan{
		Plan: planTypeOrDefault(spec.Plan, domain.SensorPlanAccelGyroStd),
	}
	pl.ID = id
	pl.CreatedAt = f.timeOrNow(spec.CreatedAt)
	pl.UpdatedAt = f.timeOrNow(spec.UpdatedAt)
	pl.DeletedAt = spec.DeletedAt
	return pl
}

func (f *Factory) Binding(instanceID, planID uuid.UUID, spec BindingSpec) *domain.SensorPlanBinding {
	binding := &domain.SensorPlanBinding{
		SensorInstanceID: instanceID,
		SensorPlanID:     planID,
	}
	binding.CreatedAt = f.timeOrNow(spec.CreatedAt)
	binding.UpdatedAt = f.timeOrNow(spec.UpdatedAt)
	binding.DeletedAt = spec.DeletedAt
	return binding
}

func (f *Factory) CertificateRequest(spec CertificateSpec, fallbackCommonName string) CertificateRequest {
	return CertificateRequest{
		CommonName: stringOrDefault(spec.CommonName, fallbackCommonName),
		DNSNames:   append([]string(nil), spec.DNSNames...),
		ValidFor:   durationOrDefault(spec.ValidFor, 365*24*time.Hour),
	}
}

func defaultName(prefix string, id uuid.UUID) string {
	shortID := strings.ReplaceAll(id.String(), "-", "")
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return prefix + "-" + shortID
}
