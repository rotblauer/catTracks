package catTracks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rotblauer/trackpoints/trackPoint"
	bolt "go.etcd.io/bbolt"
)

type timePeriodStats struct {
	TeamStats trackPoint.CatStats            `json:"team"`
	Cat       map[string]trackPoint.CatStats `json:"cat"`
}

func buildTimePeriodStats(numDays int) (stats timePeriodStats, e error) {
	d := -24 * numDays
	pts, e := getPointsSince(time.Now().Add(time.Duration(d) * time.Hour))
	if e != nil {
		fmt.Println(e)
		return stats, e
	}

	catPts := make(map[string]trackPoint.CatStats)
	for _, name := range pts.UniqueNames() { // erbody
		catPts[name] = pts.ForName(name).Statistics()
	}

	stats = timePeriodStats{
		TeamStats: pts.Statistics(),
		Cat:       catPts,
	}
	return stats, e

}

func getPointsSince(since time.Time) (trackPoint.TrackPoints, error) {

	var err error
	var points []*trackPoint.TrackPoint

	err = GetDB("master").View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(trackKey))
		c := b.Cursor()

		// TODO: fix or delete me
		min := i64tob(since.UnixNano())

		// Iterate over the 90's.
		for k, v := c.Seek(min); k != nil; k, v = c.Next() {
			var tp *trackPoint.TrackPoint
			json.Unmarshal(v, &tp)
			points = append(points, tp)
		}
		return err
	})

	return points, err
}
