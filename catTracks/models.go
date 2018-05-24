package catTracks

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
	"log"
	"os"
	"path/filepath"
	"os/user"
)

var punktlichTileDBPathRelHome = filepath.Join("punktlich.rotblauer.com", "tester.db")

type LastKnown map[string]trackPoint.TrackPoint
type Metadata struct {
	KeyN int
	KeyNUpdated time.Time
	LastUpdatedAt time.Time
	LastUpdatedBy string
	LastUpdatedPointsN int
	TileDBLastUpdated time.Time
}

func getmetadata() (out []byte, err error) {
	err = GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		out = b.Get([]byte("metadata"))
		return nil
	})
	return
}
func storemetadata(lastpoint trackPoint.TrackPoint, lenpointsupdated int) error {
	db := GetDB()
	e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))

		// if not initialized, run the stats which takes a hot second
		var keyN int

		var tileDBLastUpdated time.Time
		var homedir string
		usr, err := user.Current()
		if err != nil {
			log.Println("get current user err", err)
			homedir = os.Getenv("HOME")
		} else {
			homedir = usr.HomeDir
		}
		dbpath := filepath.Join(homedir, punktlichTileDBPathRelHome)
		dbpath = filepath.Clean(dbpath)
		log.Println("dbpath", dbpath)
		dbfi, err := os.Stat(dbpath)
		if err == nil {
			tileDBLastUpdated = dbfi.ModTime()
		} else {
			log.Println("err tile db path stat:", err)
		}

		v := b.Get([]byte("metadata"))
		md := &Metadata{}
		var keyNUpdated time.Time

		if v == nil {
			log.Println("updating bucket stats key_n...")
			keyN = tx.Bucket([]byte(trackKey)).Stats().KeyN
			log.Println("initialized metadata", "keyN:", keyN)
			keyNUpdated = time.Now().UTC()
		} else {
			if e := json.Unmarshal(v, md); e != nil {
				return e
			}
		}
		if md != nil && (md.KeyNUpdated.IsZero() || time.Since(md.KeyNUpdated) > 24 * time.Hour) {
			log.Println("updating bucket stats key_n...")
			log.Println("  because", md==nil, md.KeyNUpdated, md.KeyNUpdated.IsZero(), time.Since(md.KeyNUpdated) > 24 * time.Hour)
			keyN = tx.Bucket([]byte(trackKey)).Stats().KeyN
			log.Println("updated metadata keyN:", keyN)
			keyNUpdated = time.Now().UTC()
		} else {
			log.Println("dont update keyn", md==nil, md.KeyNUpdated, md.KeyNUpdated.IsZero(), time.Since(md.KeyNUpdated) > 24 * time.Hour)
			keyN = md.KeyN+lenpointsupdated
		}

		d := &Metadata{
			KeyN:               keyN,
			LastUpdatedAt:      time.Now().UTC(),
			LastUpdatedBy:      lastpoint.Name,
			LastUpdatedPointsN: lenpointsupdated,
			TileDBLastUpdated:  tileDBLastUpdated,
		}
		if !keyNUpdated.IsZero() {
			d.KeyNUpdated = keyNUpdated
		} else {
			d.KeyNUpdated = md.KeyNUpdated
		}
		by, e := json.Marshal(d)
		if e != nil {
			return nil
		}
		if e := b.Put([]byte("metadata"), by); e != nil {
			return e
		}

		return nil
	})
	return e
}
func getLastKnownData() (out []byte, err error) {
	err = GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		out = b.Get([]byte("lastknown"))
		return nil
	})
	return
}
func storeLastKnown(tp trackPoint.TrackPoint) {
	//lastKnownMap[tp.Name] = tp
	GetDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))

		lk := LastKnown{}

		v := b.Get([]byte("lastknown"))
		if e := json.Unmarshal(v, &lk); e != nil {
			log.Println("error unmarshalling nil lastknown", tp)
		}
		lk[tp.Name] = tp
		if by, e := json.Marshal(lk); e == nil {
			if e := b.Put([]byte("lastknown"), by); e != nil {
				return e
			}
		} else {
			log.Println("err marshalling lastknown", tp)
		}
		return nil
	})
}

func storePoints(trackPoints trackPoint.TrackPoints) error {
	var err error
	for _, point := range trackPoints {
		err = storePoint(point)
		if err != nil {
			return err
		}
	}
	if err == nil {
		l := len(trackPoints)
		err = storemetadata(trackPoints[l-1], l)
		storeLastKnown(trackPoints[l-1])
	}
	return err
}

func buildTrackpointKey(tp trackPoint.TrackPoint) []byte {
	if tp.Uuid == "" {
		if tp.ID != 0 {
			return i64tob(tp.ID)
		}
		return i64tob(tp.Time.UnixNano())
	}
	// have uuid
	k := []byte{}
	k = append(k, i64tob(tp.Time.UnixNano())...)
	k = append(k, []byte(tp.Uuid)...)
	return k
}

func storePoint(tp trackPoint.TrackPoint) error {

	var err error
	if tp.Time.IsZero() {
		tp.Time = time.Now()
	}

	GetDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))

		// id, _ := b.NextSequence()
		// trackPoint.ID = int(id)

		tp.ID = tp.Time.UnixNano() //dunno if can really get nanoy, or if will just *1000.

		key := buildTrackpointKey(tp)

		if exists := b.Get(key); exists != nil {
			// make sure it's ours
			var existingTrackpoint trackPoint.TrackPoint
			e := json.Unmarshal(exists, &existingTrackpoint)
			if e != nil {
				fmt.Println("Checking on an existing trackpoint and got an error with one of the existing trackpoints unmarshaling.")
			}
			if existingTrackpoint.Name == tp.Name {
				fmt.Println("Got redundant track; not storing: ", tp.Name, tp.Time)
				return nil
			}
		}
		// gets "" case nontestesing
		tp.Name = getTestesPrefix() + tp.Name

		trackPointJSON, err := json.Marshal(tp)
		if err != nil {
			return err
		}
		err = b.Put(key, trackPointJSON)
		if err != nil {
			fmt.Println("Didn't save post trackPoint in bolt.", err)
			return err
		}
		// p := quadtree.NewPoint(tp.Lat, tp.Lng, &tp)
		// if !GetQT().Insert(p) {
		// 	fmt.Println("Couldn't add to quadtree: ", p)
		// }
		fmt.Println("Saved trackpoint: ", tp)
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	return err
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

//TODO make queryable ala which cat when
// , channel chan *trackPoint.TrackPoint
func getPointsQT(query *query) (tps trackPoint.TPs, err error) {

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
