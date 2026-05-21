package server

import (
	"fmt"
	"io"
	"os"

	pb "geoip-receiver/proto/location"

	mqttio "github.com/eclipse/paho.mqtt.golang"
)

type LocationServer struct {
	pb.UnimplementedLocationReceiverServer
	mqtt mqttio.Client
}

func New(mqtt mqttio.Client) *LocationServer {
	return &LocationServer{mqtt: mqtt}
}

func (s *LocationServer) Stream(stream pb.LocationReceiver_StreamServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		raw := fmt.Sprintf("%s:%s:%.6f:%.6f:%d", msg.CorrelationId, msg.Id, msg.Lat, msg.Lon, msg.Timestamp)
		s.mqtt.Publish("geoip/location", 0, false, raw).Wait()
		fmt.Printf("published: correlation_id=%s id=%s lat=%.6f lon=%.6f\n", msg.CorrelationId, msg.Id, msg.Lat, msg.Lon)

		if err := stream.Send(&pb.Ack{Ok: true}); err != nil {
			fmt.Fprintf(os.Stderr, "send ack error: %v\n", err)
			return err
		}
	}
}
