package catTracks

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"path"

	"github.com/boltdb/bolt" // TOOD: use coreos
	"github.com/golang/geo/s2"
	"github.com/rotblauer/trackpoints/trackPoint"
)

var (
	db           *bolt.DB // master
	edgeDB       *bolt.DB
	trackKey     = "tracks"
	statsKey     = "stats"
	statsDataKey = "storage" // use: bucket.Put(statsDataKey, value), bucket.Get(statsDataKey)
	allBuckets   = []string{trackKey, statsKey, "names", "geohash"}
)

// GetDB is db getter.
func GetDB(name string) *bolt.DB {
	switch name {
	case "master", "":
		return db
	case "edge":
		return edgeDB
	default:
		panic("no db by that name")
	}
}

func initBuckets(db *bolt.DB, buckets []string) error {
	err := db.Update(func(tx *bolt.Tx) error {
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
	// master
	db, err = bolt.Open(path.Join("db", "tracks.db"), 0666, nil)
	if err != nil {
		fmt.Println("Could not initialize Bolt database @master. ", err)
		return err
	}
	if err := initBuckets(GetDB("master"), allBuckets); err != nil {
		fmt.Println("Err initing buckets @master.", err)
	}

	// edge
	edgeDB, err = bolt.Open(path.Join("db", "edge.db"), 0666, nil)
	if err != nil {
		fmt.Println("Could not initialize Bolt database @edge. ", err)
		return err
	}
	// only init track key for edge db
	if err := initBuckets(GetDB("edge"), []string{trackKey}); err != nil {
		return err
	}
	return nil
}

//BuildIndexBuckets populates name, lat, and long buckets from main "tracks" (time) bucket.
func BuildIndexBuckets() error {
	err := db.View(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(trackKey)).ForEach(func(key, val []byte) error {

			var tp trackPoint.TrackPoint
			if err := json.Unmarshal(val, &tp); err != nil {
				return err
			}

			// update "name"
			if err := db.Update(func(txx *bolt.Tx) error {
				bname := txx.Bucket([]byte("names"))

				bByName, _ := bname.CreateBucketIfNotExists([]byte(tp.Name))

				err := bByName.Put(buildTrackpointKey(tp), val)
				return err
			}); err != nil {
				return err
			}

			// under geohasher keys
			if err := db.Update(func(txx *bolt.Tx) error {
				b := txx.Bucket([]byte("geohash"))

				// Compute the CellID for lat, lng
				c := s2.CellIDFromLatLng(s2.LatLngFromDegrees(tp.Lat, tp.Lng))

				// store the uint64 value of c to its bigendian binary form
				hashkey := make([]byte, 8)
				binary.BigEndian.PutUint64(hashkey, uint64(c))

				e := b.Put(hashkey, val)
				if e != nil {
					fmt.Println("shit geohash index err", e)
					return fmt.Errorf("shit geohash index err: %v", e)
				}
				return nil
			}); err != nil {
				return err
			}
			return nil
		})
		return err
	})
	return err
}
