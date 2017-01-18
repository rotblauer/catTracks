package catTracks

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
	"sort"
	"time"
)

//Store a snippit of life

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func storePoints(trackPoints trackPoint.TrackPoints) error {
	var err error
	for _, point := range trackPoints {
		err = storePoint(point)
		if err != nil {
			return err
		}
	}
	return err
}

func storePoint(trackPoint trackPoint.TrackPoint) error {

	var err error
	if trackPoint.Time.IsZero() {
		trackPoint.Time = time.Now()
	}

	go func() {
		GetDB().Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(trackKey))

			id, _ := b.NextSequence()
			trackPoint.ID = int(id)

			trackPointJSON, err := json.Marshal(trackPoint)
			if err != nil {
				return err
			}
			err = b.Put(itob(trackPoint.ID), trackPointJSON)
			if err != nil {
				fmt.Println("Didn't save post trackPoint in bolt.", err)
				return err
			}
			fmt.Println("Saved trackpoint: ", trackPoint)
			return nil
		})
	}()

	if err != nil {
		fmt.Println(err)
	}
	return err
}

//get everthing in the db... can do filtering some other day

//TODO make queryable ala which cat when
func getAllPoints() (trackPoint.TrackPoints, error) {

	var err error
	var trackPoints trackPoint.TrackPoints

	err = GetDB().View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(trackKey))

		if b.Stats().KeyN > 0 {
			c := b.Cursor()
			for trackPointkey, trackPointval := c.First(); trackPointkey != nil; trackPointkey, trackPointval = c.Next() {
				//only if trackPoint is in given trackPoints key set (we don't want all trackPoints just feeded times)
				//but if no ids given, return em all
				var trackPoint trackPoint.TrackPoint
				json.Unmarshal(trackPointval, &trackPoint)
				trackPoints = append(trackPoints, trackPoint)
			}
		} else {
			//cuz its not an error if no trackPoints
			return nil
		}
		return err
	})
	sort.Sort(trackPoints) //implements interfacing facing methods from trackpoints
	return trackPoints, err
}
