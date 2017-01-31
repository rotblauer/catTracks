package catTracks

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

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
