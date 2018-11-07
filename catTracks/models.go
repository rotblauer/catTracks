package catTracks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"sync"
	"time"

	"compress/gzip"
	"github.com/boltdb/bolt"
	"github.com/kpawlik/geojson"
	"github.com/rotblauer/tileTester2/note"
	"github.com/rotblauer/trackpoints/trackPoint"
	"log"
	"os"
	"path/filepath"
)

var punktlichTileDBPathRelHome = filepath.Join("punktlich.rotblauer.com", "tester.db")

type LastKnown map[string]trackPoint.TrackPoint
type Metadata struct {
	KeyN               int
	KeyNUpdated        time.Time
	LastUpdatedAt      time.Time
	LastUpdatedBy      string
	LastUpdatedPointsN int
	TileDBLastUpdated  time.Time
}

func getmetadata() (out []byte, err error) {
	err = GetDB("master").View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		out = b.Get([]byte("metadata"))
		return nil
	})
	return
}
func storemetadata(lastpoint trackPoint.TrackPoint, lenpointsupdated int) error {
	db := GetDB("master")
	e := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))

		// if not initialized, run the stats which takes a hot second
		var keyN int

		var tileDBLastUpdated time.Time
		// var homedir string
		// usr, err := user.Current()
		// if err != nil {
		// 	log.Println("get current user err", err)
		// 	homedir = os.Getenv("HOME")
		// } else {
		// 	homedir = usr.HomeDir
		// }
		// dbpath := filepath.Join(homedir, punktlichTileDBPathRelHome)
		// dbpath = filepath.Clean(dbpath)
		dbpath := GetDB("master").Path()
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
		if md != nil && (md.KeyNUpdated.IsZero() || time.Since(md.KeyNUpdated) > 24*time.Hour) {
			log.Println("updating bucket stats key_n...")
			log.Println("  because", md == nil, md.KeyNUpdated, md.KeyNUpdated.IsZero(), time.Since(md.KeyNUpdated) > 24*time.Hour)
			keyN = 0
			// keyN = tx.Bucket([]byte(trackKey)).Stats().KeyN
			log.Println("updated metadata keyN:", keyN)
			keyNUpdated = time.Now().UTC()
		} else {
			log.Println("dont update keyn", md == nil, md.KeyNUpdated, md.KeyNUpdated.IsZero(), time.Since(md.KeyNUpdated) > 24*time.Hour)
			keyN = md.KeyN + lenpointsupdated
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
	err = GetDB("master").View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		out = b.Get([]byte("lastknown"))
		return nil
	})
	return
}

func storeLastKnown(tp trackPoint.TrackPoint) {
	//lastKnownMap[tp.Name] = tp
	lk := LastKnown{}
	if err := GetDB("master").Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))

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
	}); err != nil {
		log.Printf("error storing last known: %v", err)
	} else {
		log.Printf("stored last known: lk=%v\ntp=%v", lk, tp)
	}
}

type F struct {
	p  string // path to file
	f  *os.File
	gf *gzip.Writer
	je *json.Encoder
}

func CreateGZ(s string, compressLevel int) (f F) {
	fi, err := os.OpenFile(s, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Printf("Error in Create file\n")
		panic(err)
	}
	gf, err := gzip.NewWriterLevel(fi, compressLevel)
	if err != nil {
		log.Printf("Error in Create gz \n")
		panic(err)
	}
	je := json.NewEncoder(gf)
	f = F{s, fi, gf, je}
	return
}

func CloseGZ(f F) {
	// Close the gzip first.
	f.gf.Flush()
	f.gf.Close()
	f.f.Close()
}

func TrackToFeature(trackPointCurrent trackPoint.TrackPoint) *geojson.Feature {
	// convert to a feature
	p := geojson.NewPoint(geojson.Coordinate{geojson.Coord(trackPointCurrent.Lng), geojson.Coord(trackPointCurrent.Lat)})

	//currently need speed, name,time
	trimmedProps := make(map[string]interface{})
	trimmedProps["Speed"] = trackPointCurrent.Speed
	trimmedProps["Name"] = trackPointCurrent.Name
	trimmedProps["Time"] = trackPointCurrent.Time
	trimmedProps["UnixTime"] = trackPointCurrent.Time.Unix()
	trimmedProps["Elevation"] = trackPointCurrent.Elevation

	if ns, e := trackPointCurrent.Notes.AsNoteStructured(); e == nil {
		trimmedProps["Notes"] = ns.CustomNote
		trimmedProps["Pressure"] = ns.Pressure
		trimmedProps["Activity"] = ns.Activity
		if ns.HasValidVisit() {
			// TODO: ok to use mappy sub interface here?
			trimmedProps["Visit"] = ns.Visit
		}
	} else if nf, e := trackPointCurrent.Notes.AsFingerprint(); e == nil {
		// maybe do something with identity consolidation?
	} else {
		trimmedProps["Notes"] = trackPointCurrent.Notes.AsNoteString()
	}
	// var currentNote note.Note
	// var currentNote note.NotesField
	// e := json.Unmarshal([]byte(trackPointCurrent.Notes), &currentNote)
	// if e != nil {
	// 	trimmedProps["Notes"] = currentNote.CustomNote
	// 	trimmedProps["Pressure"] = currentNote.Pressure
	// 	trimmedProps["Activity"] = currentNote.Activity
	// } else {
	// 	trimmedProps["Notes"] = trackPointCurrent.Notes
	// }
	return geojson.NewFeature(p, trimmedProps, 1)
}

var NotifyNewEdge = make(chan bool, 1000)

var masterGZLock sync.Mutex

func storePoints(trackPoints trackPoint.TrackPoints) error {
	var err error
	// var f F
	// var fdev F
	var fedge F
	featureChan := make(chan *geojson.Feature, 100000)
	featureChanDevop := make(chan *geojson.Feature, 100000)
	featureChanEdge := make(chan *geojson.Feature, 100000)
	defer close(featureChan)
	defer close(featureChanDevop)
	defer close(featureChanEdge)
	if tracksGZPathEdge != "" {
		fedgeName := fmt.Sprintf(tracksGZPathEdge+"-wip-%d", time.Now().UnixNano())
		fedge = CreateGZ(fedgeName, gzip.BestCompression)
		go func(f F) {
			for feat := range featureChanEdge {
				if feat == nil {
					continue
				}
				f.je.Encode(feat)
			}
			CloseGZ(f)
			os.Rename(f.p, fmt.Sprintf(tracksGZPathEdge+"-fin-%d", time.Now().UnixNano()))
			NotifyNewEdge <- true
		}(fedge)
	}
	// only freya (no --proc flags, just append to master.json.gz for funsies)
	if tracksGZPath != "" && tracksGZPathEdge == "" {
		go func() {
			masterGZLock.Lock()
			defer masterGZLock.Unlock()
			mgz := CreateGZ(tracksGZPath, gzip.BestCompression)
			for feat := range featureChan {
				if feat == nil {
					continue
				}
				mgz.je.Encode(feat)
			}
			CloseGZ(mgz)
		}()
	}
	plusn := 0
	for _, point := range trackPoints {
		e := storePoint(point)
		if e != nil {
			log.Println("store point error: ", e)
			continue
		}
		plusn++
		var t2f *geojson.Feature
		if tracksGZPath != "" || tracksGZPathEdge != "" || tracksGZPathDevop != "" {
			t2f = TrackToFeature(point)
		}
		if tracksGZPath != "" {
			featureChan <- t2f
		}
		if tracksGZPathEdge != "" {
			featureChanEdge <- t2f
		}
	}
	// 47131736
	if tracksGZPath != "" {
		p := filepath.Join(filepath.Dir(tracksGZPath), "TOTALTRACKSCOUNT")
		b, be := ioutil.ReadFile(p)
		if be == nil {
			i, ie := strconv.Atoi(string(b))
			if ie == nil {
				i = i + plusn
				ioutil.WriteFile(p, []byte(strconv.Itoa(i)), 0666)
			}
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

	if tp.Lat > 90 || tp.Lat < -90 {
		return fmt.Errorf("invalid coordinate: lat=%d", tp.Lat)
	}
	if tp.Lng > 180 || tp.Lng < -180 {
		return fmt.Errorf("invalid coordinate: lng=%d", tp.Lng)
	}

	err = GetDB("master").Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))

		// Note that tp.ID is not the db key. ID is a uniq identifier per cat only.
		tp.ID = tp.Time.UnixNano() //dunno if can really get nanoy, or if will just *1000.

		key := buildTrackpointKey(tp)

		if exists := b.Get(key); exists != nil {
			// make sure same cat
			var existingTrackpoint trackPoint.TrackPoint
			e := json.Unmarshal(exists, &existingTrackpoint)
			if e != nil {
				fmt.Println("Checking on an existing trackpoint and got an error with one of the existing trackpoints unmarshaling.")
				return fmt.Errorf("unmarshal error: %v", e)
			}
			// use Name and Uuid conditions because Uuid tracking was introduced after Name, so not all points/cats/apps have it. So Name is backwards-friendly.
			if existingTrackpoint.Name == tp.Name && existingTrackpoint.Uuid == tp.Uuid {
				fmt.Println("Got redundant track; not storing: ", tp.Name, tp.Uuid, tp.Time)
				return fmt.Errorf("duplicate point")
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
			return err
		}
		fmt.Println("Saved trackpoint: ", tp)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func getAllStoredPoints() (tps trackPoint.TPs, e error) {
	start := time.Now()

	e = GetDB("master").View(func(tx *bolt.Tx) error {
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
