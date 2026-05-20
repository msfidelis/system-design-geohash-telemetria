package entities

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

type CarState struct {
	Geohash5  string
	Timestamp int64
}
