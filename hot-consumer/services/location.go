package services

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/msfidelis01/geoip-teste/hot-consumer/entities"
)

type LocationService struct {
	session *gocql.Session
}

func NewLocationService(session *gocql.Session) *LocationService {
	return &LocationService{session: session}
}

func (s *LocationService) ReadCarState(id string) (*entities.CarState, error) {
	var state entities.CarState
	err := s.session.Query(
		`SELECT geohash_5, timestamp FROM car_location_by_id WHERE id = ?`, id,
	).Scan(&state.Geohash5, &state.Timestamp)
	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read car state: %w", err)
	}
	return &state, nil
}

func (s *LocationService) UpsertLWW(event entities.LocationEvent) error {
	current, err := s.ReadCarState(event.ID)
	if err != nil {
		return err
	}

	prev := make(map[string]interface{})
	var applied bool

	if current == nil {
		applied, err = s.session.Query(
			`INSERT INTO car_location_by_id (id, timestamp, lat, lon, geohash_12, geohash_9, geohash_7, geohash_5)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS`,
			event.ID, event.Timestamp, event.Lat, event.Lon,
			event.Geohash12, event.Geohash9, event.Geohash7, event.Geohash5,
		).MapScanCAS(prev)
	} else {
		applied, err = s.session.Query(
			`UPDATE car_location_by_id
			 SET timestamp = ?, lat = ?, lon = ?, geohash_12 = ?, geohash_9 = ?, geohash_7 = ?, geohash_5 = ?
			 WHERE id = ?
			 IF timestamp < ?`,
			event.Timestamp, event.Lat, event.Lon,
			event.Geohash12, event.Geohash9, event.Geohash7, event.Geohash5,
			event.ID, event.Timestamp,
		).MapScanCAS(prev)
	}
	if err != nil {
		return fmt.Errorf("lwt car_location_by_id: %w", err)
	}
	if !applied {
		fmt.Printf("[hot] skipped stale event correlation_id=%s id=%s incoming=%d\n",
			event.CorrelationID, event.ID, event.Timestamp)
		return nil
	}

	if current != nil && current.Geohash5 != event.Geohash5 {
		if err := s.session.Query(
			`DELETE FROM car_location_by_geohash WHERE geohash_5 = ? AND id = ?`,
			current.Geohash5, event.ID,
		).Exec(); err != nil {
			return fmt.Errorf("delete old geohash bucket: %w", err)
		}
	}

	return s.session.Query(
		`INSERT INTO car_location_by_geohash (geohash_5, id, timestamp, lat, lon, geohash_12, geohash_9, geohash_7)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		event.Geohash5, event.ID, event.Timestamp, event.Lat, event.Lon,
		event.Geohash12, event.Geohash9, event.Geohash7,
	).Exec()
}
