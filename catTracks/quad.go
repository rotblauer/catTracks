package catTracks

import (
	"encoding/json"
	// "errors"
	"fmt"
	"github.com/asim/quadtree"
	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
)

var (
	qt                    *quadtree.QuadTree
	www                   quadtree.AABB
	halfwayAroundTheWorld = 20000001.42 // half world's circumference in meters
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

		count := 0 //limiter
		ver := b.ForEach(func(key, val []byte) error {
			count = count + 1

			// if count > 90 {
			// 	fmt.Println("skipping point", count)
			// 	return nil
			// } else {

			fmt.Println("adding point", count)

			var tp trackPoint.TrackPoint
			err := json.Unmarshal(val, &tp)
			// fmt.Println(tp)
			if err != nil {
				return err
			}
			p := quadtree.NewPoint(tp.Lat, tp.Lng, &tp)

			// qt.Insert(p)

			if qt.Insert(p) {
				return nil
			} else {
				fmt.Println("Couldn't insert point to quadtree.")
			}

			// err = errors.New("Could not insert point into quadtree.")
			// return err

			// }
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
