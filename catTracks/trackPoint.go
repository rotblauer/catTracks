package catTracks

import (
	"google.golang.org/appengine"
	"time"
)

type TrackPoint struct {
	LatLong   appengine.GeoPoint `json:"latLong"`
	Elevation float64            `json:"elevation"`
	HeartRate float64            `json:"heartrate"`
	Time      time.Time          `json:"time"`
	Notes     string             `json:"notes"`
}
