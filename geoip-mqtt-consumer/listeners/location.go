package listeners

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"geoip-mqtt-consumer/entities"
	"geoip-mqtt-consumer/services"

	mqttio "github.com/eclipse/paho.mqtt.golang"
)

type LocationListener struct {
	service *services.LocationService
}

func NewLocationListener(service *services.LocationService) *LocationListener {
	return &LocationListener{service: service}
}

func (l *LocationListener) Handle(_ mqttio.Client, msg mqttio.Message) {
	raw, err := parse(msg.Payload())
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		return
	}

	event := l.service.Enrich(raw)

	if err := l.service.Publish(event); err != nil {
		fmt.Fprintf(os.Stderr, "publish error: %v\n", err)
	}
}

func parse(payload []byte) (entities.RawLocation, error) {
	parts := strings.Split(string(payload), ":")
	if len(parts) != 4 {
		return entities.RawLocation{}, fmt.Errorf("expected 4 parts, got %d: %s", len(parts), payload)
	}

	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return entities.RawLocation{}, fmt.Errorf("invalid lat %q: %w", parts[1], err)
	}

	lon, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return entities.RawLocation{}, fmt.Errorf("invalid lon %q: %w", parts[2], err)
	}

	ts, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return entities.RawLocation{}, fmt.Errorf("invalid timestamp %q: %w", parts[3], err)
	}

	return entities.RawLocation{
		ID:        parts[0],
		Lat:       lat,
		Lon:       lon,
		Timestamp: ts,
	}, nil
}
