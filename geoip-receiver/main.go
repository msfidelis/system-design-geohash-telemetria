package main

import (
	"fmt"
	"net"
	"os"
	"time"

	mqttpkg "geoip-receiver/pkg/mqtt"
	pb "geoip-receiver/proto/location"
	"geoip-receiver/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func grpcAddr() string {
	if v := os.Getenv("GRPC_ADDR"); v != "" {
		return v
	}
	return ":50051"
}

func mqttAddr() string {
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		return v
	}
	return "tcp://mqtt:1883"
}

func main() {
	mqttClient, err := mqttpkg.Connect(mqttAddr(), "geoip-receiver")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mqtt connect error: %v\n", err)
		os.Exit(1)
	}
	defer mqttClient.Disconnect(250)
	fmt.Printf("connected to mqtt %s\n", mqttAddr())

	lis, err := net.Listen("tcp", grpcAddr())
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen error: %v\n", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	pb.RegisterLocationReceiverServer(srv, server.New(mqttClient))

	fmt.Printf("geoip-receiver listening on %s\n", grpcAddr())
	if err := srv.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "serve error: %v\n", err)
		os.Exit(1)
	}
}
