package grpc

import (
	"context"
	"io"
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/orchestrators/ingestor"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IngestorConsumerNoGeneric interface {
	ID() uuid.UUID
	State() ingestor.IngestStatus
}
type IngestorConsumerState interface {
	State() ingestor.IngestStatus
}

type IngestorConsumer[T any] interface {
	ports.IngestorConsumer[T, ingestor.IngestStatus]
}

type SensorMeasurementsService struct {
	proto.UnimplementedSensorServiceServer

	ingestAccelGyroStd        IngestorConsumer[*domain.AccelGyroMeasurement]
	ingestAccelGyroIndustrial IngestorConsumer[*domain.AccelGyroMeasurement]
	ingestAccelGyroRealtime   IngestorConsumer[*domain.AccelGyroMeasurement]

	ingestStates map[uuid.UUID]ingestor.IngestStatus

	checkIngestorsStatesInterval time.Duration
	stop                         chan struct{}
}

func NewSensorMeasurementsService(
	ingestAccelGyroStd IngestorConsumer[*domain.AccelGyroMeasurement],
	ingestAccelGyroIndustrial IngestorConsumer[*domain.AccelGyroMeasurement],
	ingestAccelGyroRealtime IngestorConsumer[*domain.AccelGyroMeasurement],
	checkIngestorsStatesInterval time.Duration,
) *SensorMeasurementsService {
	return &SensorMeasurementsService{
		ingestAccelGyroStd:           ingestAccelGyroStd,
		ingestAccelGyroIndustrial:    ingestAccelGyroIndustrial,
		ingestAccelGyroRealtime:      ingestAccelGyroRealtime,
		ingestStates:                 make(map[uuid.UUID]ingestor.IngestStatus, 3),
		checkIngestorsStatesInterval: checkIngestorsStatesInterval,
		stop:                         make(chan struct{}),
	}
}

func (svc *SensorMeasurementsService) CheckIngestorsStates(ctx context.Context) (stop func()) {
	ticker := time.NewTicker(svc.checkIngestorsStatesInterval)

	ingestors := []IngestorConsumerNoGeneric{
		svc.ingestAccelGyroStd,
		svc.ingestAccelGyroIndustrial,
		svc.ingestAccelGyroRealtime,
	}

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, v := range ingestors {
					svc.ingestStates[v.ID()] = v.State()
				}
			case <-ctx.Done():
				return
			case <-svc.stop:
				return
			}
		}
	}()

	return func() {
		close(svc.stop)
	}
}

func (svc *SensorMeasurementsService) StreamCustomMeasurements(srv proto.SensorService_StreamCustomMeasurementsServer) error {
	return status.Error(codes.Unimplemented, "method StreamCustomMeasurements not implemented")
}

func (svc *SensorMeasurementsService) UnaryCustomMeasurements(ctx context.Context, req *proto.CustomMeasurementsRequest) (*proto.IngestAck, error) {
	return nil, status.Error(codes.Unimplemented, "method UnaryCustomMeasurements not implemented")
}

func (svc *SensorMeasurementsService) StreamAccelGyroMeasurements(srv proto.SensorService_StreamAccelGyroMeasurementsServer) error {
	ctx := srv.Context()

	gtwctx, gtwFound := ctxvalues.GetAuthSensorGateway(ctx)
	if !gtwFound {
		return status.Error(codes.Unauthenticated, "sensor gateway not found in context")
	}

	if len(gtwctx.Sensors) == 0 {
		return status.Error(codes.FailedPrecondition, "no active sensor plans found for gateway's sensor instances")
	}

	var opts = applyDefaultsIngestAccelGyroOptions(nil)
	lastFlush := time.Now()
	count := 0
	agg := NewAckAggregator(opts.IngestOptions.AckMode)

	for {
		req, err := srv.Recv()
		if err == io.EOF {
			if err := srv.Send(agg.ToProto()); err != nil {
				return status.Errorf(codes.Internal, "final stream send error: %v", err)
			}
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "stream recv error: %v", err)
		}
		if req.Options != nil {
			opts = applyDefaultsIngestAccelGyroOptions(req.Options)
		}

		for _, sensor := range req.Sensors {
			l := len(sensor.Measurements)

			sensorID, err := uuid.Parse(sensor.SensorId)
			if err != nil {
				agg.AddRejectedN(sensor.SensorId, proto.RejectionCode_INVALID_SENSOR_ID, "invalid sensor ID format", uint64(l))
				continue
			}

			plan, planFound := gtwctx.Sensors[sensorID]
			if !planFound {
				agg.AddRejectedN(sensor.SensorId, proto.RejectionCode_UNKNOWN_SENSOR_OR_UNKNOWN_PLAN, "unknown sensor or unknown plan", uint64(l))
				continue
			}

			count += l
			agg.AddAccepted(uint64(l))

			measurements := AccelGyroDTO(sensorID, gtwctx.TenantID, sensor.Measurements)
			NormalizeAccelGyro(opts, measurements)

			switch plan {
			case domain.SensorPlanAccelGyroIndustrial:
				svc.ingestAccelGyroIndustrial.Submit(measurements)
			case domain.SensorPlanAccelGyroStd:
				svc.ingestAccelGyroStd.Submit(measurements)
			case domain.SensorPlanAccelGyroRealtime:
				svc.ingestAccelGyroRealtime.Submit(measurements)
			}
		}

		now := time.Now()
		if uint32(count) >= *opts.IngestOptions.AckEveryN ||
			now.Sub(lastFlush) > time.Duration(*opts.IngestOptions.AckMaxIntervalMs)*time.Millisecond {

			if err := srv.Send(agg.ToProto()); err != nil {
				return status.Errorf(codes.Internal, "stream send error: %v", err)
			}

			// Reset next flush
			count = 0
			lastFlush = now
			agg = NewAckAggregator(opts.IngestOptions.AckMode)
		}
	}
}

func (svc *SensorMeasurementsService) UnaryAccelGyroMeasurements(ctx context.Context, req *proto.AccelGyroMeasurementsRequest) (*proto.IngestAck, error) {
	return nil, status.Error(codes.Unimplemented, "method UnaryAccelGyroMeasurements not implemented")
}
