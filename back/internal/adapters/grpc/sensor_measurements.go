package grpc

import (
	"context"
	"io"

	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/orchestrators"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SensorMeasurementsService struct {
	proto.UnimplementedSensorServiceServer

	sensorPlanBinding ports.SensorPlanBindingOrchestrator

	ingestAccelGyroStd        *orchestrators.Ingestor[*domain.AccelGyroMeasurement]
	ingestAccelGyroIndustrial *orchestrators.Ingestor[*domain.AccelGyroMeasurement]
	ingestAccelGyroRealtime   *orchestrators.Ingestor[*domain.AccelGyroMeasurement]
}

func (svc *SensorMeasurementsService) StreamCustomMeasurements(srv proto.SensorService_StreamCustomMeasurementsServer) error {
	return status.Error(codes.Unimplemented, "method StreamCustomMeasurements not implemented")
}

func (svc *SensorMeasurementsService) UnaryCustomMeasurements(ctx context.Context, req *proto.CustomMeasurementsRequest) (*proto.IngestAck, error) {
	return nil, status.Error(codes.Unimplemented, "method UnaryCustomMeasurements not implemented")
}

func (svc *SensorMeasurementsService) StreamAccelGyroMeasurements(srv proto.SensorService_StreamAccelGyroMeasurementsServer) error {
	ctx := srv.Context()

	gtw, gtwFound := ctxvalues.GetSensorGateway(ctx)
	if !gtwFound {
		return status.Error(codes.Unauthenticated, "sensor gateway not found in context")
	}

	var instanceIDs = make([]uuid.UUID, 0, len(gtw.SensorInstances))
	for _, instances := range gtw.SensorInstances {
		instanceIDs = append(instanceIDs, instances.ID)
	}

	plans, err := svc.sensorPlanBinding.ListActiveByInstanceIDs(ctx, instanceIDs)
	if err != nil {
		return status.Error(codes.Internal, "failed to list active sensor plans")
	}

	var planByInstance = make(map[uuid.UUID]*domain.SensorPlanBinding, len(plans))
	for _, plan := range plans {
		switch plan.Plan {
		case domain.SensorPlanAccelGyroIndustrial, domain.SensorPlanAccelGyroStd, domain.SensorPlanAccelGyroRealtime:
			planByInstance[plan.SensorInstanceID] = plan
		default:
		}
	}

	for {
		_, err := srv.Recv()
		if err == io.EOF {
			_ = srv.Send(nil)
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "stream recv error: %v", err)
		}

	}
}

func (svc *SensorMeasurementsService) UnaryAccelGyroMeasurements(ctx context.Context, req *proto.AccelGyroMeasurementsRequest) (*proto.IngestAck, error) {
	return nil, status.Error(codes.Unimplemented, "method UnaryAccelGyroMeasurements not implemented")
}
