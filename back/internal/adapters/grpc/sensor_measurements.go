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

func (svc *SensorMeasurementsService) StreamMeasurements(srv proto.SensorService_StreamCustomMeasurementsServer) error {
	var accepted, rejected uint64

	for {
		req, err := srv.Recv()
		if err == io.EOF {
			_ = srv.Send(&proto.IngestAck{
				Accepted: accepted,
				Rejected: rejected,
				Status:   computeIngestStatus(accepted, rejected, false),
			})
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "stream recv error: %v", err)
		}

	}
}

func (svc *SensorMeasurementsService) UnaryMeasurements(ctx context.Context, req *proto.CustomMeasurementsRequest) (*proto.IngestAck, error) {
	return new(proto.IngestAck), nil
}

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
