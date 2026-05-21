package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	pb "github.com/msfidelis01/geoip-teste/simulacao/proto/location"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
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

func receiverAddr() string {
	if v := os.Getenv("GRPC_ADDR"); v != "" {
		return v
	}
	return "geoip-receiver:50051"
}

func runStream(client pb.LocationReceiverClient, cars []*car, rng *rand.Rand) error {
	s, err := client.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}

	fmt.Printf("stream open — simulating %d cars (Praça da Sé + MASP)\n", len(cars))

	for {
		c := cars[rng.Intn(len(cars))]
		c.move(rng)

		if err := s.Send(&pb.LocationPayload{
			Id:        c.id,
			Lat:       c.lat,
			Lon:       c.lon,
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			return fmt.Errorf("send: %w", err)
		}

		fmt.Printf("sent: %s %.6f %.6f\n", c.id, c.lat, c.lon)

		time.Sleep(time.Duration(500+rng.Intn(1500)) * time.Millisecond)
	}
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

	conn, err := grpc.NewClient(
		receiverAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewLocationReceiverClient(conn)

	fmt.Printf("connecting to geoip-receiver at %s\n", receiverAddr())

	backoff := time.Second
	for {
		if err := runStream(client, cars, rng); err != nil {
			fmt.Fprintf(os.Stderr, "stream error: %v — reconnecting in %s\n", err, backoff)
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
		} else {
			backoff = time.Second
		}
	}
}
