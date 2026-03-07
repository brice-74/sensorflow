package grpc

import (
	"context"
	"io"

	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/orchestrators"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SensorMeasurementsService struct {
	proto.UnimplementedSensorServiceServer

	ingestAccelGyroStd        *orchestrators.Ingestor[*domain.AccelGyroMeasurement]
	ingestAccelGyroIndustrial *orchestrators.Ingestor[*domain.AccelGyroMeasurement]
	ingestAccelGyroRealtime   *orchestrators.Ingestor[*domain.AccelGyroMeasurement]
}

func (svc *SensorMeasurementsService) StreamCustomMeasurements(srv proto.SensorService_StreamCustomMeasurementsServer) error {

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

func (svc *SensorMeasurementsService) UnaryCustomMeasurements(ctx context.Context, req *proto.CustomMeasurementsRequest) (*proto.IngestAck, error) {
	return nil, nil
}

func (svc *SensorMeasurementsService) StreamAccelGyroMeasurements(srv proto.SensorService_StreamAccelGyroMeasurementsServer) error {
	return status.Error(codes.Unimplemented, "method StreamAccelGyroMeasurements not implemented")
}

func (svc *SensorMeasurementsService) UnaryAccelGyroMeasurements(ctx context.Context, req *proto.AccelGyroMeasurementsRequest) (*proto.IngestAck, error) {

	return nil, status.Error(codes.Unimplemented, "method UnaryAccelGyroMeasurements not implemented")
}
