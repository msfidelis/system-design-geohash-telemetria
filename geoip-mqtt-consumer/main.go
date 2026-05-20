package main

import (
	"fmt"
	"os"

	"github.com/msfidelis01/geoip-teste/geoip-receiver/listeners"
	mqttpkg "github.com/msfidelis01/geoip-teste/geoip-receiver/pkg/mqtt"
	natspkg "github.com/msfidelis01/geoip-teste/geoip-receiver/pkg/nats"
	"github.com/msfidelis01/geoip-teste/geoip-receiver/services"
	"github.com/nats-io/nats.go/jetstream"
)

func mqttAddr() string {
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		return v
	}
	return "tcp://mqtt:1883"
}

func natsURL() string {
	if v := os.Getenv("NATS_URL"); v != "" {
		return v
	}
	return "nats://localhost:4222"
}

func main() {
	nc, err := natspkg.Connect(natsURL())
	if err != nil {
		fmt.Fprintf(os.Stderr, "nats connect error: %v\n", err)
		os.Exit(1)
	}
	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jetstream init error: %v\n", err)
		os.Exit(1)
	}

	if err := natspkg.SetupStream(js); err != nil {
		fmt.Fprintf(os.Stderr, "stream setup error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("connected to nats %s — stream GEOIP ready\n", natsURL())

	locationService := services.NewLocationService(js)
	locationListener := listeners.NewLocationListener(locationService)

	mqttClient, err := mqttpkg.Connect(mqttAddr(), "geoip-receiver", locationListener.Handle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mqtt connect error: %v\n", err)
		os.Exit(1)
	}
	defer mqttClient.Disconnect(250)

	mqttClient.Subscribe("geoip/location", 0, nil).Wait()
	fmt.Printf("connected to mqtt %s — subscribed to geoip/location\n", mqttAddr())

	select {}
}
