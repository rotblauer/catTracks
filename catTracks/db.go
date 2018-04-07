package catTracks

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"path"

	"github.com/boltdb/bolt"
	"github.com/golang/geo/s2"
	"github.com/rotblauer/trackpoints/trackPoint"
)

var (
	db         *bolt.DB
	trackKey   = "tracks"
	statsKey = "stats"
	statsDataKey = "storage" // use: bucket.Put(statsDataKey, value), bucket.Get(statsDataKey)
	allBuckets = []string{trackKey, statsKey, "names", "geohash"}
)

// GetDB is db getter.
func GetDB() *bolt.DB {
	return db
}

func initBuckets(buckets []string) error {
	err := GetDB().Update(func(tx *bolt.Tx) error {
		var e error
		for _, buck := range buckets {
			_, e = tx.CreateBucketIfNotExists([]byte(buck))
			if e != nil {
				return e
			}
		}
		return e
	})
	return err
}

// InitBoltDB sets up initial stuff, like the file and necesary buckets
func InitBoltDB() error {
	var err error
	db, err = bolt.Open(path.Join("db", "tracks.db"), 0666, nil)
	if err != nil {
		fmt.Println("Could not initialize Bolt database. ", err)
		return err
	}

	err = initBuckets(allBuckets)
	if err != nil {
		fmt.Println("Err initing buckets.", err)
	}
	return err
}

//BuildIndexBuckets populates name, lat, and long buckets from main "tracks" (time) bucket.
func BuildIndexBuckets() error {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))

		b.ForEach(func(key, val []byte) error {
			var tp trackPoint.TrackPoint
			json.Unmarshal(val, &tp)

			// update "name"
			db.Update(func(txx *bolt.Tx) error {
				bname := txx.Bucket([]byte("names"))

				bByName, _ := bname.CreateBucketIfNotExists([]byte(tp.Name))

				bByName.Put(itob(tp.ID), val)

				return nil
			})

			// under geohasher keys
			db.Update(func(txx *bolt.Tx) error {
				b := txx.Bucket([]byte("geohash"))

				// Compute the CellID for lat, lng
				c := s2.CellIDFromLatLng(s2.LatLngFromDegrees(tp.Lat, tp.Lng))

				// store the uint64 value of c to its bigendian binary form
				hashkey := make([]byte, 8)
				binary.BigEndian.PutUint64(hashkey, uint64(c))

				e := b.Put(hashkey, val)
				if e != nil {
					fmt.Println("shit geohash index err", e)
				}
				return nil
			})

			return nil
		})
		return nil
	})
	return nil
}
