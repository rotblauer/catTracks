package catTracks

import (
	"google.golang.org/appengine"
	"time"
)

// Stores a snippet of life, love, and location
type TrackPoint struct {
	Name      string             `json:"name"`
	LatLong   appengine.GeoPoint `json:"latLong"`
	Elevation float64            `json:"elevation"`
	Speed     float64            `json:"speed"`
	Tilt      float64            `json:"tilt"`
	Heading   float64            `json:"heading"`
	HeartRate float64            `json:"heartrate"`
	Time      time.Time          `json:"time"`
	Notes     string             `json:"notes"`
}
