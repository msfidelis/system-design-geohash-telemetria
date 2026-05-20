package listeners

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/msfidelis01/geoip-teste/hot-consumer/entities"
	"github.com/msfidelis01/geoip-teste/hot-consumer/services"
	"github.com/nats-io/nats.go/jetstream"
)

type LocationListener struct {
	service *services.LocationService
}

func NewLocationListener(service *services.LocationService) *LocationListener {
	return &LocationListener{service: service}
}

func (l *LocationListener) Handle(msg jetstream.Msg) {
	var event entities.LocationEvent
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal error: %v\n", err)
		msg.Ack()
		return
	}

	if err := l.service.UpsertLWW(event); err != nil {
		fmt.Fprintf(os.Stderr, "service error: %v\n", err)
		msg.Nak()
		return
	}

	fmt.Printf("[hot] id=%s geohash_5=%s geohash_7=%s lat=%.6f lon=%.6f timestamp=%d\n",
		event.ID, event.Geohash5, event.Geohash7, event.Lat, event.Lon, event.Timestamp)

	msg.Ack()
}
