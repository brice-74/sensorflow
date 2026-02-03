package grpc

import (
	"context"
	"io"

	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SensorMeasurementsService struct {
	proto.UnimplementedSensorServiceServer
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

/*
	 func computeIngestStatus(
		accepted uint64,
		rejected uint64,
		throttled bool,

	) proto.IngestStatus {
		if throttled {
			if accepted > 0 {
				return proto.IngestStatus_INGEST_STATUS_THROTTLED
			}
			return proto.IngestStatus_INGEST_STATUS_REJECTED
		}
		if accepted == 0 {
			if rejected > 0 {
				return proto.IngestStatus_INGEST_STATUS_REJECTED
			}
			return proto.IngestStatus_INGEST_STATUS_UNSPECIFIED
		}
		if rejected > 0 {
			return proto.IngestStatus_INGEST_STATUS_PARTIAL
		}
		return proto.IngestStatus_INGEST_STATUS_OK
	}
*/
func (svc *SensorMeasurementsService) StreamAccelGyroMeasurements(srv proto.SensorService_StreamAccelGyroMeasurementsServer) error {
	return status.Error(codes.Unimplemented, "method StreamAccelGyroMeasurements not implemented")
}
func (svc *SensorMeasurementsService) UnaryAccelGyroMeasurements(ctx context.Context, req *proto.AccelGyroMeasurementsRequest) (*proto.IngestAck, error) {
	return nil, status.Error(codes.Unimplemented, "method UnaryAccelGyroMeasurements not implemented")
}
