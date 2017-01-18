package catTracks

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
	"sort"
	"strconv"
	"time"
)

//Store a snippit of life

//TODO
func storePoint(trackPoint trackPoint.TrackPoint) error {

	var err error
	trackPoint.Time = time.Now()
	trackPoint.ID = strconv.Itoa(int(trackPoint.Time.UnixNano()))

	trackPointJSON, err := json.Marshal(trackPoint)
	fmt.Println("model trackPoint", trackPoint)
	fmt.Println("model jsoned trackPointJSON string:", string(trackPointJSON))
	fmt.Println("model trackPoint.ID string", string(trackPoint.ID))

	go func() {
		GetDB().Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(trackKey))
			e := b.Put([]byte(trackPoint.ID), trackPointJSON)
			if e != nil {
				fmt.Println("Didn't save post trackPoint in bolt.", e)
				return e
			}
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
