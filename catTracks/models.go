package catTracks

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/asim/quadtree"
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

// DeleteSpain deletes spain
func DeleteSpain() error {
	e := GetDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tracks"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var tp trackPoint.TrackPoint
			e := json.Unmarshal(v, &tp)
			if e != nil {
				fmt.Println("Error deleting testes.")
				return e
			}
			if tp.Lng < 12.0 && tp.Lng > -10.0 {
				b.Delete(k)
			}
		}
		return nil
	})
	return e
}

//TODO make queryable ala which cat when
func getAllPoints(query *query) ([]*trackPoint.TrackPoint, error) {

	var err error
	var coords []simpleline.Point

	if query != nil && query.IsBounded() {
		//build aabb rect
		var center = make(map[string]float64)
		//not totally sure what halfpoint means but best guess
		center["lat"] = (query.Bounds.NorthEastLat + query.Bounds.SouthWestLat) / 2.0
		center["lng"] = (query.Bounds.NorthEastLng + query.Bounds.SouthWestLng) / 2.0
		cp := quadtree.NewPoint(center["lat"], center["lng"], nil)
		half := trackPoint.Distance(center["lat"], center["lng"], center["lat"], query.Bounds.NorthEastLng)
		hp := cp.HalfPoint(half)
		ab := quadtree.NewAABB(cp, hp)
		//res = GetQT.Search(aabb)
		qres := GetQT().Search(ab)
		//for range res = coords append res[i].data
		//TODO check gainst other query params
		for _, val := range qres {
			coords = append(coords, val.Data().(*trackPoint.TrackPoint))
		}
	} else {

		//TODO make this not what it was

		err = GetDB().View(func(tx *bolt.Tx) error {
			var err error
			b := tx.Bucket([]byte(trackKey))

			// can swap out for- eacher if we figure indexing, or even want it
			b.ForEach(func(trackPointKey, trackPointVal []byte) error {

				var trackPointCurrent trackPoint.TrackPoint
				json.Unmarshal(trackPointVal, &trackPointCurrent)

				coords = append(coords, &trackPointCurrent)
				return nil

			})

			return err
		})

	}

	//? but why is there a null pointer error? how is the func being passed a nil query?
	var epsilon float64
	if query != nil {
		epsilon = query.Epsilon // just so we can separate incoming queryEps and wiggled-to Eps
	} else {
		epsilon = DefaultEpsilon // to default const
	}
	//simpleify line
	// results, sErr := simpleline.RDP(coords, 5, simpleline.Euclidean, true)
	originalCount := len(coords)
	results, err := simpleline.RDP(coords, epsilon, simpleline.Euclidean, true) //0.001 bring a 5700pt run to prox 300 (.001 scale is lat and lng)
	if err != nil {
		fmt.Println("Errrrrrr", err)
		results = coords // return coords, err //better dan nuttin //but not sure want to return the err...
	}
	fmt.Println("eps: ", epsilon)
	fmt.Println("  results: ", len(results))

	var l int
	if query != nil {
		l = query.Limit
	} else {
		l = DefaultLimit
	}
	//just a hacky shot at wiggler. pointslimits -> query eventually?
	for len(results) > l {
		epsilon = epsilon + epsilon/(1-epsilon)
		fmt.Println("wiggling eps: ", epsilon)
		//or could do with results stead of coords?
		results, err = simpleline.RDP(coords, epsilon, simpleline.Euclidean, true)
		if err != nil {
			fmt.Println("Errrrrrr", err)
			results = coords
			continue
		}
		fmt.Println("  results: ", len(results))
	}

	var tps trackPoint.TPs
	for _, insult := range results {
		o, ok := insult.(*trackPoint.TrackPoint)
		if !ok {
			fmt.Println("shittt notok")
		}
		tps = append(tps, o)
	}

	fmt.Println("Serving points.")
	fmt.Println("Original total points: ", originalCount)
	fmt.Println("post-RDP-wiggling point: ", len(results))

	sort.Sort(tps)

	return tps, err
}
