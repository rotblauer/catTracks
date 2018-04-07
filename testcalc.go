package main

import (
	"github.com/rotblauer/catTracks/catTracks"
	"log"
	"fmt"
)

func main() {
	// Open Bolt DB.
	// catTracks.InitBoltDB()
	if bolterr := catTracks.InitBoltDB(); bolterr == nil {
		defer catTracks.GetDB().Close()
	}

	if e := catTracks.CalculateAndStoreStats(3); e != nil {
		log.Println("calcstats err:", e)
		return
	}
	val, e := catTracks.GetStats()
	if e != nil {
		log.Println("getstats err:", e)
		return
	}

	fmt.Println(string(val))
}
