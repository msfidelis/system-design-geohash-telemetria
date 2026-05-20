package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	maxRadiusM = 5000.0
	stepM      = 30.0
)

// Praça da Sé
const (
	seBaseLat = -23.55028
	seBaseLon = -46.63389
)

// MASP
const (
	maspBaseLat = -23.587416
	maspBaseLon = -46.657634
)

var uuids_se = []string{
	"a1b2c3d4-0001-4e5f-8a9b-000000000001",
	"a1b2c3d4-0002-4e5f-8a9b-000000000002",
	"a1b2c3d4-0003-4e5f-8a9b-000000000003",
	"a1b2c3d4-0004-4e5f-8a9b-000000000004",
	"a1b2c3d4-0005-4e5f-8a9b-000000000005",
	"a1b2c3d4-0006-4e5f-8a9b-000000000006",
	"a1b2c3d4-0007-4e5f-8a9b-000000000007",
	"a1b2c3d4-0008-4e5f-8a9b-000000000008",
	"a1b2c3d4-0009-4e5f-8a9b-000000000009",
	"a1b2c3d4-0010-4e5f-8a9b-000000000010",
}

var uuids_masp = []string{
	"b2c3d4e5-0001-4f6a-9b0c-100000000001",
	"b2c3d4e5-0002-4f6a-9b0c-100000000002",
	"b2c3d4e5-0003-4f6a-9b0c-100000000003",
	"b2c3d4e5-0004-4f6a-9b0c-100000000004",
	"b2c3d4e5-0005-4f6a-9b0c-100000000005",
	"b2c3d4e5-0006-4f6a-9b0c-100000000006",
	"b2c3d4e5-0007-4f6a-9b0c-100000000007",
	"b2c3d4e5-0008-4f6a-9b0c-100000000008",
	"b2c3d4e5-0009-4f6a-9b0c-100000000009",
	"b2c3d4e5-0010-4f6a-9b0c-100000000010",
}

type car struct {
	id      string
	lat     float64
	lon     float64
	baseLat float64
	baseLon float64
}

func metersToDegreesLat(m float64) float64 {
	return m / 111320.0
}

func metersToDegreesLon(m float64, lat float64) float64 {
	return m / (111320.0 * math.Cos(lat*math.Pi/180.0))
}

func distanceM(lat1, lon1, lat2, lon2 float64) float64 {
	dlat := (lat2 - lat1) * 111320.0
	dlon := (lon2 - lon1) * 111320.0 * math.Cos(lat1*math.Pi/180.0)
	return math.Sqrt(dlat*dlat + dlon*dlon)
}

func (c *car) move(rng *rand.Rand) {
	angle := rng.Float64() * 2 * math.Pi
	newLat := c.lat + metersToDegreesLat(stepM*math.Sin(angle))
	newLon := c.lon + metersToDegreesLon(stepM*math.Cos(angle), c.lat)

	if distanceM(c.baseLat, c.baseLon, newLat, newLon) <= maxRadiusM {
		c.lat = newLat
		c.lon = newLon
	}
}

func brokerAddr() string {
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		return v
	}
	return "tcp://mqtt:1883"
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	cars := make([]*car, 0, len(uuids_se)+len(uuids_masp))
	for _, id := range uuids_se {
		cars = append(cars, &car{id: id, lat: seBaseLat, lon: seBaseLon, baseLat: seBaseLat, baseLon: seBaseLon})
	}
	for _, id := range uuids_masp {
		cars = append(cars, &car{id: id, lat: maspBaseLat, lon: maspBaseLon, baseLat: maspBaseLat, baseLon: maspBaseLon})
	}

	opts := mqtt.NewClientOptions().
		AddBroker(brokerAddr()).
		SetClientID("simulacao").
		SetConnectRetry(true).
		SetConnectRetryInterval(3 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Fprintf(os.Stderr, "connect error: %v\n", token.Error())
		os.Exit(1)
	}
	defer client.Disconnect(250)

	fmt.Printf("connected to %s — simulating %d cars (Praça da Sé + MASP)\n", brokerAddr(), len(cars))

	for {
		c := cars[rng.Intn(len(cars))]
		c.move(rng)

		payload := fmt.Sprintf("%s:%.6f:%.6f:%d", c.id, c.lat, c.lon, time.Now().UnixMilli())
		client.Publish("geoip/location", 0, false, payload).Wait()

		fmt.Printf("published: %s\n", payload)

		time.Sleep(time.Duration(500+rng.Intn(1500)) * time.Millisecond)
	}
}
