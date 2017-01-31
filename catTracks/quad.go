package catTracks

import (
	"encoding/json"
	"time"
	// "errors"
	"fmt"

	"github.com/asim/quadtree"
	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
)

var (
	qt                    *quadtree.QuadTree
	www                   quadtree.AABB
	halfwayAroundTheWorld = 40000001.42 // half world's circumference in meters
	centerOfTheWorld      = map[string]float64{"lat": 0.0, "lng": 0.0}
)

// GetQT eturn QT, just like DB
func GetQT() *quadtree.QuadTree {
	return qt
}

func initQTBounds() *quadtree.AABB {
	// built new centerpoint
	p := quadtree.NewPoint(centerOfTheWorld["lat"], centerOfTheWorld["lng"], nil)
	// get half point
	half := p.HalfPoint(halfwayAroundTheWorld)
	// build new AABB
	return quadtree.NewAABB(p, half)
}

//InitQT initializes quadtree by iterating through all points and inserting them into in-memory (yikes!) qt
func InitQT() error {
	var e error

	//tinker with default qt sizing
	quadtree.MaxDepth = 24
	quadtree.Capacity = 72 //lets really blow it up

	//being new quadtree
	fmt.Println("initing qt...")
	qt = quadtree.New(initQTBounds(), 0, nil)

	//stick points into quadtree
	e = GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))

		ver := b.ForEach(func(key, val []byte) error {

			var tp trackPoint.TrackPoint
			err := json.Unmarshal(val, &tp)
			// fmt.Println(tp)
			if err != nil {
				return err
			}
			p := quadtree.NewPoint(tp.Lat, tp.Lng, &tp)

			if qt.Insert(p) {
				return nil
			} else {
				fmt.Println("Couldn't insert point to quadtree.")
			}
			return nil
		})
		return ver
	})
	if e != nil {
		fmt.Println(e)
	}

	if e == nil {
		fmt.Println("Successfully added all trackpoints to quadtree.")
	}

	return e
}

func getPointsFromQT(query *query) (tps trackPoint.TPs) {
	start := time.Now()

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
	//for range res = tps append res[i].data
	//TODO check gainst other query params
	// fmt.Println("server quad res length: ", len(qres))
	for _, val := range qres {
		tps = append(tps, val.Data().(*trackPoint.TrackPoint))
	}

	fmt.Printf("Found %s points with quadtree method - %s\n", len(qres), time.Since(start))

	return tps
}
