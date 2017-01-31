package catTracks

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/geo/s2"
	"github.com/rotblauer/trackpoints/trackPoint"
)

// NewKeyWithInt returns a key prefixed by 'K' with value i
func NewKeyWithInt(id int64) []byte {
	key := bytes.NewBufferString("K")
	binary.Write(key, binary.BigEndian, id)

	return key.Bytes()
}

// NewGeoKey generates a new key using a position & a key
// place + time + person
func NewGeoKey(tp trackPoint.TrackPoint) []byte {
	t := tp.Time
	kint := t.UnixNano()
	kid := NewKeyWithInt(kint)

	c := s2.CellIDFromLatLng(s2.LatLngFromDegrees(tp.Lat, tp.Lng)) //geohash.EncodeWithPrecision(latitude, longitude, 6)
	placeKey := make([]byte, 8)
	binary.BigEndian.PutUint64(placeKey, uint64(c))

	personKey := []byte("N" + tp.Name)

	zkey := append(placeKey, kid...)
	zkey = append(zkey, personKey...)

	// K<unixnanotime><cellIdBytes>N<personName>
	return zkey
}

// GeoKeyPrefix return prefixes to lookup using a GeoKey and timerange
// func GeoKeyPrefix(start, stop time.Time) []string {
// 	var res []string
// 	d := 10 * time.Minute
// 	var t time.Time
// 	t = start
// 	for {
// 		if t.After(stop) {
// 			break
// 		}

// 		key := "G" + t.Format("2006010215") + fmt.Sprintf("%02d", t.Minute()-t.Minute()%10)
// 		res = append(res, key)
// 		t = t.Add(d)
// 	}
// 	return res
// }

// http://blog.nobugware.com/post/2016/geo_db_s2_geohash_database/
// citiesInCellID looks for cities inside c
func citiesInCellID(c s2.CellID) (tps trackPoint.TPs) {
	// compute min & max limits for c
	bmin := make([]byte, 8)
	bmax := make([]byte, 8)
	binary.BigEndian.PutUint64(bmin, uint64(c.RangeMin()))
	binary.BigEndian.PutUint64(bmax, uint64(c.RangeMax()))

	// perform a range lookup in the DB from bmin key to bmax key, cur is our DB cursor
	// var cell s2.CellID
	GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("geohash"))
		cur := b.Cursor()
		for k, v := cur.Seek(bmin); k != nil && bytes.Compare(k, bmax) <= 0; k, v = cur.Next() {

			var tp trackPoint.TrackPoint
			json.Unmarshal(v, &tp)

			tps = append(tps, &tp)
		}
		return nil
	})
	return tps
}

//TODO make queryable ala which cat when
// , channel chan *trackPoint.TrackPoint
func getPointsGH(query *query) (tps trackPoint.TPs) {

	start := time.Now()

	// for now, use _way back_ and _now_ for time bounds... can queryfy em later
	// d := time.Duration(-10) * time.Minute
	// geoPrefixs := GeoKeyPrefix(time.Now().UTC().Add(d), time.Now().UTC())

	rect := s2.RectFromLatLng(s2.LatLngFromDegrees(query.Bounds.SouthWestLat, query.Bounds.SouthWestLng))
	rect = rect.AddPoint(s2.LatLngFromDegrees(query.Bounds.NorthEastLat, query.Bounds.NorthEastLng))

	rc := &s2.RegionCoverer{MaxLevel: 22, MaxCells: 8}
	r := s2.Region(rect.CapBound())
	covering := rc.Covering(r)

	for _, c := range covering {
		tps = append(tps, citiesInCellID(c)...)
	}

	fmt.Println("Found ", len(tps), " points with Geohash method in", time.Since(start))

	return tps
}
