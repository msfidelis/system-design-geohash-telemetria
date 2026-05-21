package services

import (
	"fmt"

	"app/entities"

	"github.com/gocql/gocql"
	"github.com/mmcloughlin/geohash"
)

type LocationService struct {
	session *gocql.Session
}

func NewLocationService(session *gocql.Session) *LocationService {
	return &LocationService{session: session}
}

func (s *LocationService) GetCarByID(id string) (*entities.Car, error) {
	var car entities.Car
	err := s.session.Query(
		`SELECT id, lat, lon, geohash_12, geohash_9, geohash_7, geohash_5, timestamp
		 FROM car_location_by_id WHERE id = ?`, id,
	).Scan(&car.ID, &car.Lat, &car.Lon, &car.Geohash12, &car.Geohash9, &car.Geohash7, &car.Geohash5, &car.Timestamp)
	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get car by id: %w", err)
	}
	return &car, nil
}

func (s *LocationService) GetCarsByRegion(lat, lon float64) (*entities.CarsByRegionResponse, error) {
	gh5 := geohash.EncodeWithPrecision(lat, lon, 5)

	iter := s.session.Query(
		`SELECT id, lat, lon, geohash_12, geohash_9, geohash_7, timestamp
		 FROM car_location_by_geohash WHERE geohash_5 = ?`, gh5,
	).Iter()

	var cars []entities.Car
	var car entities.Car
	for iter.Scan(&car.ID, &car.Lat, &car.Lon, &car.Geohash12, &car.Geohash9, &car.Geohash7, &car.Timestamp) {
		car.Geohash5 = gh5
		cars = append(cars, car)
		car = entities.Car{}
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("get cars by region: %w", err)
	}

	if cars == nil {
		cars = []entities.Car{}
	}

	return &entities.CarsByRegionResponse{
		Geohash5: gh5,
		Total:    len(cars),
		Carros:   cars,
	}, nil
}
