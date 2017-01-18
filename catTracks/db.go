package catTracks

import (
	"path"
	"github.com/boltdb/bolt"
	"fmt"
	"github.com/rotblauer/trackpoints/trackPoint"
)


var (
	db *bolt.DB
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
			_, e := tx.CreateBucketIfNotExists([]byte(trackKey))
			if e != nil {
				return e
			}
			return e
		})
	}
	return err
	// return GetDB()
}

//Store a snippit of life

//TODO
func storePoint(trackPoint trackPoint.TrackPoint) error {

	return nil
}



//get everthing in the db... can do filtering some other day

//TODO
func getAllPoints(catQ string) []trackPoint.TrackPoint {


	return nil
}
