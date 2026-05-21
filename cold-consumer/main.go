package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type locationEvent struct {
	CorrelationID string  `json:"correlation_id"`
	ID            string  `json:"id"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Geohash12     string  `json:"geohash_12"`
	Geohash9      string  `json:"geohash_9"`
	Geohash7      string  `json:"geohash_7"`
	Geohash5      string  `json:"geohash_5"`
	Timestamp     int64   `json:"timestamp"`
}

func natsURL() string {
	if v := os.Getenv("NATS_URL"); v != "" {
		return v
	}
	return "nats://localhost:4222"
}

func main() {
	nc, err := nats.Connect(natsURL(),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(3*time.Second),
	)
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

	consumer, err := js.CreateOrUpdateConsumer(context.Background(), "GEOIP", jetstream.ConsumerConfig{
		Durable:   "cold-storage",
		AckPolicy: jetstream.AckExplicitPolicy,
		// DeliverPolicy: jetstream.DeliverNewPolicy,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "consumer create error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("connected to %s — consuming stream GEOIP as cold-storage\n", natsURL())

	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		var event locationEvent
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			fmt.Fprintf(os.Stderr, "unmarshal error: %v\n", err)
			// msg.Nak()
			return
		}

		fmt.Printf("[cold] correlation_id=%s id=%s lat=%.6f lon=%.6f geohash_12=%s geohash_9=%s geohash_7=%s geohash_5=%s timestamp=%d\n",
			event.CorrelationID, event.ID, event.Lat, event.Lon, event.Geohash12, event.Geohash9, event.Geohash7, event.Geohash5, event.Timestamp)

		msg.Ack()
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "consume error: %v\n", err)
		os.Exit(1)
	}
	defer cc.Stop()

	select {}
}
