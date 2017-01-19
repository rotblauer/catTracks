package catTracks

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/deet/simpleline"
	"github.com/rotblauer/trackpoints/trackPoint"
	"sort"
	"strings"
	"time"
)

const (
	testesPrefix = "testes-------"
)

var testes = false

// SetTestes run
func SetTestes(flagger bool) {
	testes = flagger
}
func getTestesPrefix() string {
	if testes {
		return testesPrefix
	}
	return ""
}

//Store a snippit of life

// itob returns an 8-byte big endian representation of v.
func itob(v int64) []byte {
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

func storePoint(tp trackPoint.TrackPoint) error {

	var err error
	if tp.Time.IsZero() {
		tp.Time = time.Now()
	}

	go func() {
		GetDB().Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(trackKey))

			// id, _ := b.NextSequence()
			// trackPoint.ID = int(id)
			tp.ID = tp.Time.UnixNano() //dunno if can really get nanoy, or if will just *1000.
			if exists := b.Get(itob(tp.ID)); exists != nil {
				// make sure it's ours
				var existingTrackpoint trackPoint.TrackPoint
				e := json.Unmarshal(exists, &existingTrackpoint)
				if e != nil {
					fmt.Println("Checking on an existing trackpoint and got an error with one of the existing trackpoints unmarshaling.")
				}
				if existingTrackpoint.Name == tp.Name {
					fmt.Println("Got that trackpoint already. Breaking.")
					return nil
				}
			}
			// gets "" case nontestesing
			tp.Name = getTestesPrefix() + tp.Name

			trackPointJSON, err := json.Marshal(tp)
			if err != nil {
				return err
			}
			err = b.Put(itob(tp.ID), trackPointJSON)
			if err != nil {
				fmt.Println("Didn't save post trackPoint in bolt.", err)
				return err
			}
			fmt.Println("Saved trackpoint: ", tp)
			return nil
		})
	}()

	if err != nil {
		fmt.Println(err)
	}
	return err
}

// DeleteTestes wipes the entire database of all points with names prefixed with testes prefix. Saves an rm keystorke
func DeleteTestes() error {
	e := GetDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var tp trackPoint.TrackPoint
			e := json.Unmarshal(v, &tp)
			if e != nil {
				fmt.Println("Error deleting testes.")
				return e
			}
			if strings.HasPrefix(tp.Name, testesPrefix) {
				b.Delete(k)
			}
		}
		return nil
	})
	return e
}

//get everthing in the db... can do filtering some other day

//TODO make queryable ala which cat when
func getAllPoints() ([]*trackPoint.TrackPoint, error) {

	var err error
	// var trackPoints trackPoint.TrackPoints
	var coords []simpleline.Point

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
				// trackPoints = append(trackPoints, trackPoint)

				//rdp
				coords = append(coords, &trackPoint) //filler up

			}

		} else {
			//cuz its not an error if no trackPoints
			return nil
		}
		return err
	})

	originalCount := len(coords)

	//simpleify line
	// results, sErr := simpleline.RDP(coords, 5, simpleline.Euclidean, true)
	results, sErr := simpleline.RDP(coords, 0.001, simpleline.Euclidean, true) //0.001 bring a 3000pt run to prox 300 (cuz scale is lat and lng)
	if sErr != nil {
		fmt.Println("Errrrrrr", sErr)
	}

	rdpCount := len(results)

	//dis shit is fsck but fsckit
	//truncater
	// trackPoints = trackPoints[len(trackPoints)-3:] // lets go crazy with 3
	var tps trackPoint.TPs

	for _, insult := range results {

		// fmt.Println(insult)
		o, ok := insult.(*trackPoint.TrackPoint)
		if !ok {
			fmt.Println("shittt notok")
		}
		tps = append(tps, o)
	}
	fmt.Println("Serving points. Original count was ", originalCount, " and post-RDP is ", rdpCount)
	sort.Sort(tps)

	return tps, err
}
