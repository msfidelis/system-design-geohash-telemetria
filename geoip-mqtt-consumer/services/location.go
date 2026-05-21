package services

import (
	"context"
	"encoding/json"
	"fmt"

	"geoip-mqtt-consumer/entities"

	"github.com/mmcloughlin/geohash"
	"github.com/nats-io/nats.go/jetstream"
)

type LocationService struct {
	js jetstream.JetStream
}

func NewLocationService(js jetstream.JetStream) *LocationService {
	return &LocationService{js: js}
}

func (s *LocationService) Enrich(raw entities.RawLocation) entities.LocationEvent {
	return entities.LocationEvent{
		CorrelationID: raw.CorrelationID,
		ID:            raw.ID,
		Lat:           raw.Lat,
		Lon:           raw.Lon,
		Geohash12:     geohash.EncodeWithPrecision(raw.Lat, raw.Lon, 12),
		Geohash9:      geohash.EncodeWithPrecision(raw.Lat, raw.Lon, 9),
		Geohash7:      geohash.EncodeWithPrecision(raw.Lat, raw.Lon, 7),
		Geohash5:      geohash.EncodeWithPrecision(raw.Lat, raw.Lon, 5),
		Timestamp:     raw.Timestamp,
	}
}

func (s *LocationService) Publish(event entities.LocationEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if _, err := s.js.Publish(context.Background(), "geoip.location", payload); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}

	fmt.Printf("published: correlation_id=%s id=%s\n", event.CorrelationID, event.ID)
	return nil
}
