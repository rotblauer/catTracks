package catTracks

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
	"strconv"
	"time"
)

// OVERALL: per day/ per ever/ per week/ ... per time
// OVERALL: per cat/ per catTeam

// OVERALL: calculations:
// - max, min, mean, mode, median

// OVERALL: units
// - elevation, speed, distance,

//TODO per catName string
func getPointsSince(since time.Time) (trackPoint.TrackPoints, error) {

	var err error
	var points []trackPoint.TrackPoint
	sinceNano := since.UnixNano()

	err = GetDB().View(func(tx *bolt.Tx) error {
		var err error
		c := tx.Bucket([]byte(trackKey)).Cursor()

		min := []byte(strconv.Itoa(int(sinceNano)))

		// Iterate over the 90's.
		for k, v := c.Seek(min); k != nil; k, v = c.Next() {
			var tp trackPoint.TrackPoint
			json.Unmarshal(v, &tp)
			points = append(points, tp)
		}
		return err
	})

	return points, err
}
