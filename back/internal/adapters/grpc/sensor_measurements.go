package grpc

import (
	"context"
	"io"

	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SensorMeasurementsService struct {
	proto.UnimplementedSensorServiceServer
}

func (svc *SensorMeasurementsService) StreamMeasurements(srv proto.SensorService_StreamMeasurementsServer) error {
	for {
		req, err := srv.Recv()

		if err == io.EOF {
			return srv.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "stream recv error: %v", err)
		}
	}
}

func (svc *SensorMeasurementsService) UnaryMeasurements(ctx context.Context, req *proto.SensorMeasurementsRequest) (*emptypb.Empty, error) {
	return new(emptypb.Empty), nil
}
