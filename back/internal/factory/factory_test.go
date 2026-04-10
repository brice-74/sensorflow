package factory_test

import (
	"testing"
	"time"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/factory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFactoryTenantDefaultValues(t *testing.T) {
	f := factory.New(factory.WithNow(func() time.Time {
		return time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	}))

	tenant := f.Tenant(factory.TenantSpec{})

	require.NotEqual(t, uuid.Nil, tenant.ID)
	require.Equal(t, "tenant-"+tenant.ID.String()[:8], tenant.Name)
	require.Equal(t, time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC), tenant.CreatedAt)
	require.Equal(t, tenant.CreatedAt, tenant.UpdatedAt)
}

func TestFactoryExplicitOverrides(t *testing.T) {
	gatewayID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	planID := uuid.MustParse("66666666-7777-8888-9999-aaaaaaaaaaaa")
	fixedTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	firmware := "v1.2.3"
	name := "gw-prod"
	planType := domain.SensorPlanAccelGyroRealtime

	f := factory.New(factory.WithNow(func() time.Time {
		return time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	}))

	gateway := f.Gateway(gatewayID, factory.GatewaySpec{
		ID:        &gatewayID,
		Name:      &name,
		Firmware:  &firmware,
		CreatedAt: &fixedTime,
		UpdatedAt: &fixedTime,
	})
	plan := f.Plan(factory.PlanSpec{
		ID:        &planID,
		Plan:      &planType,
		CreatedAt: &fixedTime,
		UpdatedAt: &fixedTime,
	})

	require.Equal(t, gatewayID, gateway.ID)
	require.Equal(t, name, gateway.Name)
	require.Equal(t, firmware, *gateway.Firmware)
	require.Equal(t, fixedTime, gateway.CreatedAt)
	require.Equal(t, fixedTime, gateway.UpdatedAt)
	require.Equal(t, planID, plan.ID)
	require.Equal(t, planType, plan.Plan)
	require.Equal(t, fixedTime, plan.CreatedAt)
	require.Equal(t, fixedTime, plan.UpdatedAt)
}

func TestFactoryDemoSetupBuildsConsistentRelations(t *testing.T) {
	f := factory.New(factory.WithNow(func() time.Time {
		return time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	}))

	setup := f.DemoSetup(factory.DemoSetupSpec{
		Tenant: factory.TenantSpec{},
		Gateways: []factory.GatewaySpec{
			{},
			{},
		},
		InstancesPerGateway: 2,
	})

	require.NotNil(t, setup.Tenant)
	require.Len(t, setup.Gateways, 2)
	require.Len(t, setup.Instances, 4)
	require.Len(t, setup.Bindings, 4)
	require.Len(t, setup.Plans, 3)
	require.Len(t, setup.Certificates, 2)

	for i, gateway := range setup.Gateways {
		require.Equal(t, setup.Tenant.ID, gateway.TenantID)
		require.NotEqual(t, uuid.Nil, gateway.ID)
		require.Equal(t, gateway.Name, setup.Certificates[i].Request.CommonName)
	}

	for _, instance := range setup.Instances {
		require.NotEqual(t, uuid.Nil, instance.ID)
		require.NotNil(t, instance.ActiveSensorPlanBinding)
		require.Equal(t, instance.ID, instance.ActiveSensorPlanBinding.SensorInstanceID)
		require.NotEqual(t, uuid.Nil, instance.ActiveSensorPlanBinding.SensorPlanID)
		require.Equal(t, instance.ActiveSensorPlanBinding.SensorPlanID, instance.ActiveSensorPlanBinding.Plan.ID)
	}
}
