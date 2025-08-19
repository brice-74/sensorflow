package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/brice-74/sensorflow/cmd/ingestion/app"
	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
	"github.com/brice-74/sensorflow/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

func main() {
	cfg, err := app.ParseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse config error:", err)
		os.Exit(1)
	}

	logger, flushSentry, err := app.PrepareLogger(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		os.Exit(1)
	}
	defer flushSentry()

	logger = logger.With(
		log.Tags{"api_identifier": cfg.Instance.Identifier},
	).(log.FiberLoggerInterface)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		os.Exit(1)
	}

	s := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	proto.RegisterSensorServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
		os.Exit(1)
	}

}

type server struct {
	proto.UnimplementedSensorServiceServer
}

func (s *server) StreamMeasurements(stream proto.SensorService_StreamMeasurementsServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		fmt.Println("Pas de metadata reçu")
	}

	fmt.Println("md --- ", md)

	client := ParseDN(md["x-client-dn"][0])
	fmt.Println(client)

	for {
		req, err := stream.Recv()
		if err != nil {
			fmt.Println("break", err)
			break
		}
		fmt.Println("receive request", req.String())
	}

	fmt.Println("close")
	return stream.SendAndClose(nil)
}

type ClientDN struct {
	CN string
	OU string
	O  string
	L  string
	ST string
	C  string
}

func ParseDN(dn string) ClientDN {
	parts := strings.Split(dn, ",")
	result := ClientDN{}

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch key {
		case "CN":
			result.CN = val
		case "OU":
			result.OU = val
		case "O":
			result.O = val
		case "L":
			result.L = val
		case "ST":
			result.ST = val
		case "C":
			result.C = val
		}
	}

	return result
}
