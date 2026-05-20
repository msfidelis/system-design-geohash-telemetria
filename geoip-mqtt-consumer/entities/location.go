package entities

// RawLocation é o payload cru recebido do MQTT no formato id:lat:lon:timestamp
type RawLocation struct {
	ID        string
	Lat       float64
	Lon       float64
	Timestamp int64
}

// LocationEvent é o evento enriquecido com geohashes publicado no NATS
type LocationEvent struct {
	ID        string  `json:"id"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Geohash12 string  `json:"geohash_12"`
	Geohash9  string  `json:"geohash_9"`
	Geohash7  string  `json:"geohash_7"`
	Geohash5  string  `json:"geohash_5"`
	Timestamp int64   `json:"timestamp"`
}
