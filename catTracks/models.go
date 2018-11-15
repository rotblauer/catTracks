package catTracks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"sync"
	"time"

	"compress/gzip"
	"github.com/boltdb/bolt"
	"github.com/davecgh/go-spew/spew"
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

type QueryFilterPlaces struct {
	Names []string `schema:"names"`

	// start,end x arrive,depart,report
	StartArrivalT   time.Time `schema:"startArrivalT"`
	EndArrivalT     time.Time `schema:"endArrivalT"`
	StartDepartureT time.Time `schema:"startDepartureT"`
	EndDepartureT   time.Time `schema:"endDepartureT"`
	StartReportedT  time.Time `schema:"startReportedT"`
	EndReportedT    time.Time `schema:"endReportedT"`

	ReverseChrono bool `schema:"rc"` // when true, oldest first; default to newest first

	// paginatables
	StartIndex int `schema:"startI"` // 0 indexed;
	EndIndex   int `schema:"endI"`   // diff end-start = per/pagination lim

	// for geo rect bounding, maybe
	LatMin *float64 `schema:"latmin"`
	LatMax *float64 `schema:"latmax"`
	LngMin *float64 `schema:"lngmin"`
	LngMax *float64 `schema:"lngmax"`

	IncludeStats bool `schema:"stats"`
}

// var DefaultQFP = QueryFilterPlaces{
// 	Names:    []string{},
// 	EndIndex: 30, // given zero values, reverse and StartIndex=0, this returns 30 most recent places
// 	LatMin:   math.MaxFloat64,
// 	LatMax:   math.MaxFloat64,
// 	LngMin:   math.MaxFloat64,
// 	LatMax:   math.MaxFloat64,
// }

// // ByTime implements Sort interface for NoteVisit
// type ByTime []note.NoteVisit

// func (bt ByTime) Len() int {
// 	return len(bt)
// }

// func (bt ByTime) Swap(i, j int) {
// 	bt[i], bt[j] = bt[j], bt[i]
// }

// // Less compares ARRIVALTIME. This might need to be expanded or differentiated.
// func (bt ByTime) Less(i, j int) bool {
// 	return bt[i].ArrivalTime.Before(bt[j].ArrivalTime)
// }

type VisitsResponse struct {
	Visits    []*note.NoteVisit `json:"visits"`
	Stats     bolt.BucketStats  `json"bucketStats"`
	StatsTook time.Duration     `json:"statsTook"` // how long took to get bucket stats (for 10mm++ points, long time)
	Scanned   int               `json:"scanned"`   // num visits checked before mtaching filters
	Matches   int               `json:"matches"`   // num visits matching before paging/index filters
}

// btw places are actually visits. fucked that one up.
func getPlaces(qf QueryFilterPlaces) (out []byte, err error) {
	// TODO
	// - wire to router with query params
	// - filter during key iter
	// - sortable interface places
	// - places to json, new type in Note

	log.Println("handling visits q:", spew.Sdump(qf))

	var res = VisitsResponse{}
	var visits = []*note.NoteVisit{}
	var scannedN, matchingN int // nice convenience returnables for query stats, eg. matched 4/14 visits, querier can know this

	err = GetDB("master").View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(placesKey))

		if qf.IncludeStats {
			t1 := time.Now()
			res.Stats = b.Stats()
			res.StatsTook = time.Since(t1)
		}

		if b == nil {
			return fmt.Errorf("no places bolt bucket exists")
		}
		c := b.Cursor()

		// unclosed forloop conditions
		k, vis := c.First()
		condit := func() bool {
			return k != nil
		}

		// use fancy seekers/condits if reported time query parameter comes thru
		// when there's a lot of visits, this might be moar valuable
		// Note that reported time is trackpoint time, and that's the key we range over. Same key as tp.
		if !qf.StartReportedT.IsZero() {
			k, vis = c.Seek(i64tob(qf.StartReportedT.UnixNano()))
		}
		if !qf.EndReportedT.IsZero() {
			condit = func() bool {
				// i64tob uses big endian with 8 bytes
				return k != nil && bytes.Compare(k[:8], i64tob(qf.EndReportedT.UnixNano())) <= 0
			}
		}

	ITERATOR:
		for ; condit(); k, vis = c.Next() {

			scannedN++

			var nv = &note.NoteVisit{}

			err := json.Unmarshal(vis, nv)
			if err != nil {
				log.Println("error unmarshalling visit for query:", err)
				continue
			}

			// filter: names
			if len(qf.Names) > 0 || nv.Name == "" {
				// fuck... gotta x-reference tp to check cat names
				var tp = &trackPoint.TrackPoint{}

				bt := tx.Bucket([]byte(trackKey))
				tpv := bt.Get(k)
				if tpv == nil {
					log.Println("no trackpoint stored for visit:", nv)
					continue
				}
				err = json.Unmarshal(tpv, tp)
				if err != nil {
					log.Println("err unmarshalling tp for visit query:", err)
					continue
				}
				nv.Name = tp.Name
			}
			if len(qf.Names) > 0 {
				var ok bool
				for _, n := range qf.Names {
					if n == nv.Name {
						ok = true
						break
					}
				}
				if !ok {
					// doesn't match any of the given whitelisted names
					continue ITERATOR
				}
			}

			// filter: start/endT
			if !qf.StartArrivalT.IsZero() {
				if nv.ArrivalTime.Before(qf.StartArrivalT) {
					continue ITERATOR
				}
			}
			if !qf.StartDepartureT.IsZero() {
				if nv.DepartureTime.Before(qf.StartDepartureT) {
					continue ITERATOR
				}
			}
			if !qf.EndArrivalT.IsZero() {
				if nv.ArrivalTime.After(qf.EndArrivalT) {
					continue ITERATOR
				}
			}
			if !qf.EndDepartureT.IsZero() {
				if nv.DepartureTime.After(qf.EndDepartureT) {
					continue ITERATOR
				}
			}

			// filter: lat,lng x min,max
			if qf.LatMin != nil {
				if nv.PlaceParsed.Lat < *qf.LatMin {
					continue ITERATOR
				}
			}
			if qf.LatMax != nil {
				if nv.PlaceParsed.Lat > *qf.LatMax {
					continue ITERATOR
				}
			}
			if qf.LngMax != nil {
				if nv.PlaceParsed.Lng > *qf.LngMax {
					continue ITERATOR
				}
			}
			if qf.LngMin != nil {
				if nv.PlaceParsed.Lng < *qf.LngMin {
					continue ITERATOR
				}
			}

			matchingN++
			visits = append(visits, nv)
		}
		return nil
	})

	// filter: handle reverse Chrono
	if qf.ReverseChrono {
		// sort with custom Less function in closure
		sort.Slice(visits, func(i, j int) bool {
			return visits[i].ArrivalTime.Before(visits[j].ArrivalTime)
		})
	} else {
		sort.Slice(visits, func(i, j int) bool {
			return visits[i].ArrivalTime.After(visits[j].ArrivalTime)
		})
	}

	// filter: paginate with indexes, limited oob's
	// FIXME this might not even be right, just tryna avoid OoB app killers (we could theor allow negs, with fance reversing wrapping, but tldd)
	if qf.EndIndex == 0 || qf.EndIndex > len(visits) || qf.EndIndex < 0 {
		qf.EndIndex = len(visits)
	}
	if qf.StartIndex > len(visits) {
		qf.StartIndex = len(visits)
	}
	if qf.StartIndex < 0 {
		qf.StartIndex = 0
	}

	res.Visits = visits[qf.StartIndex:qf.EndIndex]
	res.Matches = matchingN
	res.Scanned = scannedN

	out, err = json.Marshal(res)

	return
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

func (f F) JE() *json.Encoder {
	return f.je
}

func CloseGZ(f F) {
	// Close the gzip first.
	f.gf.Flush()
	f.gf.Close()
	f.f.Close()
}

// func TrackToFeaturePlace(tp trackPoint.TrackPoint) *geojson.Feature {
// 	p := geojson.NewPoint(geojson.Coordinate{geojson.Coord(tp.Lng), geojson.Coord{tp.Lat}})

// 	ns, err := note.NotesField(tp.Notes).AsNoteStructured()
// 	if err != nil {
// 		return nil
// 	}

// 	if !ns.HasValidVisit() {
// 		return nil
// 	}

// 	visit, err := ns.Visit.AsVisit()
// 	if err != nil {
// 		return nil
// 	}

// 	place, err := visit.Place.AsPlace()
// 	if err != nil {
// 		return nil
// 	}

// 	props := make(map[string]interface{})
// 	props["CatName"] = tp.Name
// 	props["ArrivalTime"] = visit.ArrivalTime
// 	props["DepartureTime"] = visit.DepartureTime
// 	props["Identity"] = place.Identity
// 	props["Address"] = place.Address
// 	props["Activity"] = ns.Activity
// 	return geojson.NewFeature(p, props, 1)
// }

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

	if ns, e := note.NotesField(trackPointCurrent.Notes).AsNoteStructured(); e == nil {
		trimmedProps["Notes"] = ns.CustomNote
		trimmedProps["Pressure"] = ns.Pressure
		trimmedProps["Activity"] = ns.Activity
		if ns.HasValidVisit() {
			// TODO: ok to use mappy sub interface here?
			trimmedProps["Visit"] = ns.Visit
		}
	} else if _, e := note.NotesField(trackPointCurrent.Notes).AsFingerprint(); e == nil {
		// maybe do something with identity consolidation?
	} else {
		trimmedProps["Notes"] = note.NotesField(trackPointCurrent.Notes).AsNoteString()
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

func TrackToPlace(tp trackPoint.TrackPoint, visit note.NoteVisit) *geojson.Feature {
	p := geojson.NewPoint(geojson.Coordinate{geojson.Coord(visit.PlaceParsed.Lng), geojson.Coord(visit.PlaceParsed.Lat)})

	props := make(map[string]interface{})
	props["Name"] = tp.Name
	props["ReportedTime"] = tp.Time
	props["ArrivalTime"] = visit.ArrivalTime
	props["DepartureTime"] = visit.DepartureTime
	props["PlaceIdentity"] = visit.PlaceParsed.Identity
	props["PlaceAddress"] = visit.PlaceParsed.Address
	props["Accuracy"] = visit.PlaceParsed.Acc

	return geojson.NewFeature(p, props, 1)
}

var NotifyNewEdge = make(chan bool, 1000)
var NotifyNewPlace = make(chan bool, 1000)
var FeaturePlaceChan = make(chan *geojson.Feature, 100000)

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
	// // tracksGzpath only cuz too lazy to add another flag for places, and we'll use the tracsgz path dir
	// if tracksGZPathEdge != "" && placesLayer {
	// 	go func() {
	// 		PlacesGZLock.Lock()
	// 		defer PlacesGZLock.Unlock()
	// 		pgz := CreateGZ(filepath.Join(filepath.Dir(tracksGZPathEdge), "places.json.gz"), gzip.BestCompression)
	// 		for feat := range featurePlaceChan {
	// 			if feat == nil {
	// 				continue
	// 			}
	// 			pgz.je.Encode(feat)
	// 		}
	// 		CloseGZ(pgz)
	// 		NotifyNewPlace <- true
	// 	}()
	// }
	plusn := 0
	for _, point := range trackPoints {
		visit, e := storePoint(point)
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
		// tp has note has visit
		if !visit.ReportedTime.IsZero() && placesLayer {
			FeaturePlaceChan <- TrackToPlace(point, visit)
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

func storePoint(tp trackPoint.TrackPoint) (note.NoteVisit, error) {
	var err error
	var visit note.NoteVisit
	if tp.Time.IsZero() {
		tp.Time = time.Now()
	}

	if tp.Lat > 90 || tp.Lat < -90 {
		return visit, fmt.Errorf("invalid coordinate: lat=%.14f", tp.Lat)
	}
	if tp.Lng > 180 || tp.Lng < -180 {
		return visit, fmt.Errorf("invalid coordinate: lng=%.14f", tp.Lng)
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

		// handle storing place
		ns, err := note.NotesField(tp.Notes).AsNoteStructured()
		if err != nil {
			return nil
		}
		if !ns.HasValidVisit() {
			return nil
		}
		visit, err = ns.Visit.AsVisit()
		if err != nil {
			return nil
		}

		visit.Name = tp.Name
		visit.PlaceParsed = visit.Place.MustAsPlace()
		visit.ReportedTime = tp.Time
		visit.Duration = visit.GetDuration()

		visitJSON, err := json.Marshal(visit)
		if err != nil {
			return err
		}

		pb := tx.Bucket([]byte(placesKey))
		err = pb.Put(key, visitJSON)
		if err != nil {
			return err
		}
		fmt.Println("Saved visit: ", visit)

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return visit, err
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
