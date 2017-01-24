package catTracks

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
	"path"
)

var (
	db       *bolt.DB
	trackKey = "tracks"
)

// GetDB is db getter.
func GetDB() *bolt.DB {
	return db
}

// InitBoltDB sets up initial stuff, like the file and necesary buckets
func InitBoltDB() error {
	//sec := setting.Cfg.Section("server")
	//p := sec.Key("APP_DATA_PATH").String()
	where := path.Join("db", "tracks.db")

	var err error
	db, err = bolt.Open(where, 0666, nil)

	// return err
	if err != nil {
		fmt.Println("Could not initialize Bolt database. ", err)
	} else {
		fmt.Println("Bolt db is initialized.")
		db.Update(func(tx *bolt.Tx) error {
			// "tracks" -- this is the default bucket, keyed on time.UnixNano
			_, e := tx.CreateBucketIfNotExists([]byte(trackKey))
			if e != nil {
				return e
			}
			_, e = tx.CreateBucketIfNotExists([]byte("names"))
			if e != nil {
				return e
			}
			return e
		})
	}
	return err
	// return GetDB()
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

			return nil
		})
		return nil
	})
	return nil
}
