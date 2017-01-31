package catTracks

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/asim/quadtree"
	"github.com/boltdb/bolt"
	"github.com/deet/simpleline"
	"github.com/golang/geo/s2"
	"github.com/rotblauer/trackpoints/trackPoint"
)

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
			p := quadtree.NewPoint(tp.Lat, tp.Lng, &tp)
			if !GetQT().Insert(p) {
				fmt.Println("Couldn't add to quadtree: ", p)
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
		// could send channeler??
		tps = append(tps, o)
	}

	fmt.Println("Serving points.")
	fmt.Println("Original total points: ", originalCount)
	fmt.Println("post-RDP-wiggling point: ", len(results))

	sort.Sort(tps)

	return tps, err
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

func getAllStoredPoints() (tps trackPoint.TPs, e error) {
	start := time.Now()

	e = GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))

		// can swap out for- eacher if we figure indexing, or even want it
		b.ForEach(func(trackPointKey, trackPointVal []byte) error {

			var trackPointCurrent trackPoint.TrackPoint
			err := json.Unmarshal(trackPointVal, &trackPointCurrent)
			if err != nil {
				return err
			}

			tps = append(tps, &trackPointCurrent)
			return nil
		})
		return nil
	})
	fmt.Printf("Found %d points with iterator method - %s\n", len(tps), time.Since(start))

	return tps, e
}

// nope nope nope nope, hm. TODO.
// scale comes from ass/leaf-socket.js  getZoomLevel()
func getEpsFromScale(scale float64) float64 {
	n := math.Pow((scale / math.Pi), -1) // ~ 3.34916212
	fmt.Println("Scale", scale, " yields eps ", n)
	return n
}

func limitTrackPoints(query *query, tps trackPoint.TPs) (limitedTps trackPoint.TPs, e error) {

	start := time.Now()
	originalPointsCount := len(tps)

	var tpsSimple []simpleline.Point
	for _, tp := range tps {
		tpsSimple = append(tpsSimple, tp)
	}

	var epsilon = query.Epsilon
	// if query.Scale > 0.1 {
	// 	epsilon = getEpsFromScale(query.Scale)
	// }

	res, e := simpleline.RDP(tpsSimple, epsilon, simpleline.Euclidean, true)
	if e != nil {
		fmt.Println(e)
		return tps, e
	}

	for len(res) > query.Limit {
		epsilon = epsilon + epsilon/(1-epsilon)

		// could be rdp-ing the already rdp-ed?
		res2, e := simpleline.RDP(tpsSimple, epsilon, simpleline.Euclidean, true)
		if e != nil {
			fmt.Println("Error wiggling epsy.", e)
			res = tpsSimple
			continue
		} else {
			res = res2
		}
	}

	for _, simpleP := range res {
		tp, ok := simpleP.(*trackPoint.TrackPoint)
		if !ok {
			fmt.Println("shittt notok")
		}
		// could send channeler??
		limitedTps = append(limitedTps, tp)
	}

	fmt.Println("Limited points ", originalPointsCount, " to ", len(limitedTps), " in ", time.Since(start))
	return limitedTps, e
}

//TODO make queryable ala which cat when
// , channel chan *trackPoint.TrackPoint
func socketPointsByQueryQuadtree(query *query) (tps trackPoint.TPs, err error) {

	if query == nil {
		query = NewQuery()
	}

	query.SetDefaults() // eps, lim  catches empty vals

	if query.IsBounded() {
		tps = getPointsFromQT(query)
	} else {
		tps, err = getAllStoredPoints()
		if err != nil {
			return nil, err
		}
	}

	if len(tps) > query.Limit {
		limitedTPs, err := limitTrackPoints(query, tps)
		if err != nil {
			fmt.Println(err)
			return tps, err
		}
		tps = limitedTPs
	}

	sort.Sort(tps)

	return tps, err
}
