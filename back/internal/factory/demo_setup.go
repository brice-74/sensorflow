package factory

import (
	"github.com/brice-74/sensorflow/internal/core/domain"
)

type DemoSetupSpec struct {
	Tenant              TenantSpec
	Gateways            []GatewaySpec
	InstancesPerGateway int
	Plans               []PlanSpec
}

type DemoSetup struct {
	Tenant       *domain.Tenant
	Gateways     []*domain.SensorGateway
	Instances    []*domain.SensorInstance
	Plans        []*domain.SensorPlan
	Bindings     []*domain.SensorPlanBinding
	Certificates []*GatewayCertificate
}

func (f *Factory) DemoSetup(spec DemoSetupSpec) *DemoSetup {
	tenant := f.Tenant(spec.Tenant)
	plans := f.demoPlans(spec.Plans)

	gatewaysSpec := spec.Gateways
	if len(gatewaysSpec) == 0 {
		gatewaysSpec = []GatewaySpec{{}}
	}

	instancesPerGateway := spec.InstancesPerGateway
	if instancesPerGateway <= 0 {
		instancesPerGateway = 1
	}

	gateways := make([]*domain.SensorGateway, 0, len(gatewaysSpec))
	instances := make([]*domain.SensorInstance, 0, len(gatewaysSpec)*instancesPerGateway)
	bindings := make([]*domain.SensorPlanBinding, 0, len(gatewaysSpec)*instancesPerGateway)
	certificates := make([]*GatewayCertificate, 0, len(gatewaysSpec))

	for gatewayIndex, gatewaySpec := range gatewaysSpec {
		gateway := f.Gateway(tenant.ID, gatewaySpec)
		gateways = append(gateways, gateway)

		for instanceIndex := 0; instanceIndex < instancesPerGateway; instanceIndex++ {
			plan := plans[(gatewayIndex*instancesPerGateway+instanceIndex)%len(plans)]
			instance := f.Instance(gateway.ID, InstanceSpec{})
			binding := f.Binding(instance.ID, plan.ID, BindingSpec{})

			binding.Plan = *plan
			instance.ActiveSensorPlanBinding = binding

			instances = append(instances, instance)
			bindings = append(bindings, binding)
		}

		certSpec := gatewaySpec.Certificate
		if certSpec == nil {
			certSpec = &CertificateSpec{}
		}
		certificates = append(certificates, &GatewayCertificate{
			GatewayID: gateway.ID,
			Request:   f.CertificateRequest(*certSpec, gateway.Name),
		})
	}

	return &DemoSetup{
		Tenant:       tenant,
		Gateways:     gateways,
		Instances:    instances,
		Plans:        plans,
		Bindings:     bindings,
		Certificates: certificates,
	}
}

func (f *Factory) demoPlans(specs []PlanSpec) []*domain.SensorPlan {
	if len(specs) == 0 {
		defaults := []domain.SensorPlanType{
			domain.SensorPlanAccelGyroStd,
			domain.SensorPlanAccelGyroIndustrial,
			domain.SensorPlanAccelGyroRealtime,
		}
		plans := make([]*domain.SensorPlan, 0, len(defaults))
		for _, planType := range defaults {
			pt := planType
			plans = append(plans, f.Plan(PlanSpec{Plan: &pt}))
		}
		return plans
	}

	plans := make([]*domain.SensorPlan, 0, len(specs))
	for _, spec := range specs {
		plans = append(plans, f.Plan(spec))
	}
	return plans
}
