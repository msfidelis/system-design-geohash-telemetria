package main

import (
	"context"
	"fmt"
	"os"

	"github.com/msfidelis01/geoip-teste/hot-consumer/listeners"
	cassandrapkg "github.com/msfidelis01/geoip-teste/hot-consumer/pkg/cassandra"
	natspkg "github.com/msfidelis01/geoip-teste/hot-consumer/pkg/nats"
	"github.com/msfidelis01/geoip-teste/hot-consumer/services"
	"github.com/nats-io/nats.go/jetstream"
)

func natsURL() string {
	if v := os.Getenv("NATS_URL"); v != "" {
		return v
	}
	return "nats://localhost:4222"
}

func cassandraHost() string {
	if v := os.Getenv("CASSANDRA_HOST"); v != "" {
		return v
	}
	return "localhost"
}

func main() {
	session := cassandrapkg.Connect(cassandraHost())
	defer session.Close()
	fmt.Printf("connected to cassandra %s\n", cassandraHost())

	nc, err := natspkg.Connect(natsURL())
	if err != nil {
		fmt.Fprintf(os.Stderr, "nats connect error: %v\n", err)
		os.Exit(1)
	}
	defer nc.Drain()
	fmt.Printf("connected to nats %s\n", natsURL())

	locationService := services.NewLocationService(session)
	locationListener := listeners.NewLocationListener(locationService)

	js, err := jetstream.New(nc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jetstream init error: %v\n", err)
		os.Exit(1)
	}

	consumer, err := js.CreateOrUpdateConsumer(context.Background(), "GEOIP", jetstream.ConsumerConfig{
		Durable:   "hot-storage",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "consumer create error: %v\n", err)
		os.Exit(1)
	}

	cc, err := consumer.Consume(locationListener.Handle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "consume error: %v\n", err)
		os.Exit(1)
	}
	defer cc.Stop()

	fmt.Println("consuming stream GEOIP as hot-storage")
	select {}
}
