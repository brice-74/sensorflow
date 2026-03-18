package grpc

import "github.com/brice-74/sensorflow/internal/adapters/grpc/proto"

func applyDefaultsIngestAccelGyroOptions(opts *proto.IngestAccelGyroOptions) *proto.IngestAccelGyroOptions {
	if opts == nil {
		opts = &proto.IngestAccelGyroOptions{}
	}

	opts.IngestOptions = applyDefaultsIngestOptions(opts.IngestOptions, 50, 1000, 3000)

	if opts.AccelUnit == nil {
		opts.AccelUnit = proto.AccelUnit_M_S2.Enum()
	}
	if opts.GyroUnit == nil {
		opts.GyroUnit = proto.GyroUnit_DEG_S.Enum()
	}
	if opts.TemperatureUnit == nil {
		opts.TemperatureUnit = proto.TemperatureUnit_CELSIUS.Enum()
	}
	return opts
}

func applyDefaultsIngestOptions(opts *proto.IngestOptions, maxRejectedDetails, ackEveryN, ackMaxIntervalMs uint32) *proto.IngestOptions {
	if opts == nil {
		opts = &proto.IngestOptions{}
	}
	if opts.AckMode == proto.AckMode_ACK_MODE_UNSPECIFIED {
		opts.AckMode = proto.AckMode_ACK_MODE_FULL
	}
	if opts.MaxRejectedDetails == nil {
		opts.MaxRejectedDetails = &maxRejectedDetails
	}
	if opts.AckEveryN == nil {
		opts.AckEveryN = &ackEveryN
	}
	if opts.AckMaxIntervalMs == nil {
		opts.AckMaxIntervalMs = &ackMaxIntervalMs
	}
	return opts
}
