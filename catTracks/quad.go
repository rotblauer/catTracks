package catTracks

import (
	"encoding/json"
	"errors"
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

	//being new quadtree
	fmt.Println("initing qt...")
	qt = quadtree.New(initQTBounds(), 0, nil)

	// //stick points into quadtree
	// e = GetDB().View(func(tx *bolt.Tx) error {
	// 	b := tx.Bucket([]byte(trackKey))
	// 	ver := b.ForEach(func(key, val []byte) error {
	// 		var tp trackPoint.TrackPoint
	// 		err := json.Unmarshal(val, &tp)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		p := quadtree.NewPoint(tp.Lat, tp.Lng, tp)
	// 		if qt.Insert(p) {
	// 			return err
	// 		}
	// 		err = errors.New("Could not insert point into quadtree.")
	// 		return err
	// 	})
	// 	return ver
	// })
	// if e == nil {
	// 	fmt.Println("Successfully added all trackpoints to quadtree.")
	// }

	return e
}
