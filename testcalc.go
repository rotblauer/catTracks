package main

import (
	"github.com/rotblauer/catTracks/catTracks"
	"log"
	"time"
	"github.com/boltdb/bolt"
	"encoding/json"
	"github.com/rotblauer/trackpoints/trackPoint"
	"bytes"
	"fmt"
)

func main() {
	// Open Bolt DB.
	// catTracks.InitBoltDB()
	if bolterr := catTracks.InitBoltDB(); bolterr == nil {
		defer catTracks.GetDB().Close()
	}

	// clear out old stats from dev db.
	catTracks.GetDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("stats"))
		e := b.ForEach(func(k, v []byte) error {
			if err := b.Delete(k); err != nil {
				return err
			}
			return nil
		})
		return e
	})



	//l, err := time.LoadLocation("America/New_York")
	//if err != nil {
	//	log.Fatal(err)
	//}
	refT := time.Date(2018, time.March, 1, 12, 0, 0, 0, time.UTC)
	log.Println("now   ", catTracks.I64tob(time.Now().UnixNano()))
	log.Println("reft  ", catTracks.I64tob(refT.UnixNano()))

	var RFC3339NanoFormatNormal = "2006-01-02T15:04:05.000000000Z07:00"
	var seekMeTime time.Time
	catTracks.GetDB().Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("tracks")).Cursor()
		ob, e := tx.CreateBucketIfNotExists([]byte("tracksNatural"))
		if e != nil {
			panic(e)
		}
		var n = 0


		for k, v := c.First(); k != nil; k, v = c.Next() {
			var trackPointCurrent trackPoint.TrackPoint
			err := json.Unmarshal(v, &trackPointCurrent)
			if err != nil {
				return err
			}
			var normalKey = []byte{}
			normalKey = trackPointCurrent.Time.AppendFormat(normalKey, RFC3339NanoFormatNormal)

			if e := ob.Put(normalKey, v); e != nil {
				panic(e)
			}

			if n == 500 {
				seekMeTime = trackPointCurrent.Time
			}

			if n > 1000 {
				break
			}
			n++
		}

		return nil
	})
	catTracks.GetDB().View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("tracksNatural")).Cursor()
		log.Println("seek time", seekMeTime)
		var seekStart = []byte(seekMeTime.Format(RFC3339NanoFormatNormal))
		var seekEnd = []byte(seekMeTime.Add(500*time.Hour).Format(RFC3339NanoFormatNormal))
		log.Println("seek start", seekStart)
		var nn = 0
		for k, v := c.Seek(seekStart); k != nil && bytes.Compare(k, seekEnd) <= 0; k, v = c.Next() {
			if nn == 0 {
				log.Println("sook", k)
			}
			fmt.Printf("%s: %s\n", k, v)
		}
		return nil
	})


	//if e := catTracks.CalculateAndStoreStatsByDateAndSpanStepping(refT, 1*time.Hour, -120*24*time.Hour); e != nil {
	//	log.Println("calcstats err:", e)
	//	return
	//}
	//val, e := catTracks.GetStats(refT.Add(4*time.Hour), -16*time.Hour)
	////val, e := catTracks.GetStats(time.Now(), -2455*time.Hour)
	//if e != nil {
	//	log.Println("getstats err:", e)
	//	return
	//}
	//
	//fmt.Println(string(val))
}
