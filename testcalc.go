package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/rotblauer/catTracks/catTracks"
	"log"
	"time"
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

	l, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}
	refT := time.Date(2018, time.April, 1, 12, 0, 0, 0, l)
	if e := catTracks.CalculateAndStoreStatsByDateAndSpanStepping(refT, 1*time.Hour, -12*time.Hour); e != nil {
		log.Println("calcstats err:", e)
		return
	}
	val, e := catTracks.GetStats(refT.Add(4*time.Hour), -16*time.Hour)
	//val, e := catTracks.GetStats(time.Now(), -2455*time.Hour)
	if e != nil {
		log.Println("getstats err:", e)
		return
	}

	fmt.Println(string(val))
}
