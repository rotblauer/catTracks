package catTracks

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rotblauer/trackpoints/trackPoint"
	bolt "go.etcd.io/bbolt"
)

func ExportPostGIS() {
	GetDB("master").View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var track trackPoint.TrackPoint
			err := json.Unmarshal(v, &track)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Println(track.Name)
		}
		return nil
	})

}
