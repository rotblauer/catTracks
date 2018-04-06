package catTracks

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/deet/simpleline"
	"github.com/golang/geo/s2"
	"github.com/rotblauer/trackpoints/trackPoint"
)

// http://blog.nobugware.com/post/2016/geo_db_s2_geohash_database/
// citiesInCellID looks for cities inside c
func citiesInCellID(c s2.CellID) []simpleline.Point {
	// compute min & max limits for c
	bmin := make([]byte, 8)
	bmax := make([]byte, 8)
	binary.BigEndian.PutUint64(bmin, uint64(c.RangeMin()))
	binary.BigEndian.PutUint64(bmax, uint64(c.RangeMax()))

	var coords []simpleline.Point

	// perform a range lookup in the DB from bmin key to bmax key, cur is our DB cursor
	var cell s2.CellID
	GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("geohash"))
		cur := b.Cursor()
		for k, v := cur.Seek(bmin); k != nil && bytes.Compare(k, bmax) <= 0; k, v = cur.Next() {

			buf := bytes.NewReader(k)
			binary.Read(buf, binary.BigEndian, &cell)

			var tp trackPoint.TrackPoint
			json.Unmarshal(v, &tp)
			//fsckerr

			coords = append(coords, &tp)

			// // Read back a city
			// ll := cell.LatLng()
			// lat := float64(ll.Lat.Degrees())
			// lng := float64(ll.Lng.Degrees())
			// name = string(v)
			// fmt.Println(lat, lng, name)

		}
		return nil
	})
	return coords
}

//TODO make queryable ala which cat when
// , channel chan *trackPoint.TrackPoint
func socketPointsByQueryGeohash(query *query) (trackPoint.TPs, error) {

	var err error
	var coords []simpleline.Point

	if query != nil && query.IsBounded() {

		start := time.Now()

		rect := s2.RectFromLatLng(s2.LatLngFromDegrees(query.Bounds.SouthWestLat, query.Bounds.SouthWestLng))
		rect = rect.AddPoint(s2.LatLngFromDegrees(query.Bounds.NorthEastLat, query.Bounds.NorthEastLng))

		rc := &s2.RegionCoverer{MaxLevel: 20, MaxCells: 8}
		r := s2.Region(rect.CapBound())
		covering := rc.Covering(r)

		for _, c := range covering {
			// citiesInCellID(c)
			coords = append(coords, citiesInCellID(c)...)
		}

		fmt.Printf("Found %d points with Geohash method  - %s", len(coords), time.Since(start))

	} else {

		start := time.Now()
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

		fmt.Printf("Found %d points with iterator method - %s", len(coords), time.Since(start))
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

	var l int
	if query != nil {
		l = query.Limit
	} else {
		l = DefaultLimit
	}
	//just a hacky shot at wiggler. pointslimits -> query eventually?
	var results []simpleline.Point
	if originalCount > l {

		results, err = simpleline.RDP(coords, epsilon, simpleline.Euclidean, true) //0.001 bring a 5700pt run to prox 300 (.001 scale is lat and lng)
		if err != nil {
			fmt.Println("Errrrrrr", err)
			results = coords // return coords, err //better dan nuttin //but not sure want to return the err...
		}
		fmt.Println("eps: ", epsilon)
		fmt.Println("  results: ", len(results))

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

	} else {
		results = coords
	}

	var tps trackPoint.TPs
	for _, insult := range results {
		o, ok := insult.(*trackPoint.TrackPoint)
		if !ok {
			fmt.Println("shittt notok")
		}
		// could send channeler??
		tps = append(tps, o)

	}

	fmt.Println("Serving points.")
	fmt.Println("Original total points: ", originalCount)
	fmt.Println("post-RDP-wiggling point: ", len(results))

	sort.Sort(tps)

	return tps, err
}
