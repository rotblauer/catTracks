package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/rotblauer/catTracks/catTracks"
)

//start the url handlers, special init for everything?

//Toodle to do , Command line port arg, might mover er to main
func main() {
	var porty int          // set port to listen on
	var clearDBTestes bool // clear test points from db and returns
	var testesRun bool     // sets prefix for incoming points
	var buildIndexes bool  // builds names and geohash indexes and returns
	var searchType string  // sets search type to 'geohash' or 'quadtree' (qt default)

	flag.IntVar(&porty, "port", 8080, "port to serve and protect")
	flag.BoolVar(&clearDBTestes, "castrate-first", false, "clear out db of testes prefixed points") //TODO clear only certain values, ie prefixed with testes based on testesRun
	flag.BoolVar(&testesRun, "testes", false, "testes run prefixes name with testes-")              //hope that's your phone's name
	flag.BoolVar(&buildIndexes, "build-indexes", false, "build index buckets for original trackpoints")
	flag.StringVar(&searchType, "search", "quadtree", "set search type to either 'quadtree' or 'geohash'")

	flag.Parse()

	// Open Bolt DB.
	// catTracks.InitBoltDB()
	if bolterr := catTracks.InitBoltDB(); bolterr == nil {
		defer catTracks.GetDB().Close()
	}

	if clearDBTestes {
		e := catTracks.DeleteTestes()
		if e != nil {
			log.Println(e)
		}
		return
	}

	if buildIndexes {
		catTracks.BuildIndexBuckets() //cleverly always returns nil
		return
	}

	catTracks.SetSearch(searchType)

	// if using quadtree (Default) search type, set quadtree into mem (into mem for now)
	fmt.Println("Using search type: ", catTracks.GetSearchType())
	if searchType == "quadtree" {
		if qterr := catTracks.InitQT(); qterr != nil {
			log.Println("Error initing QT.")
			log.Println(qterr)
		}
	}

	catTracks.InitMelody() // sets up websockets, see ./socket.go

	// set env things
	catTracks.SetTestes(testesRun) //is false defaulter, false prefixes names with ""
	catTracks.SetSearch(searchType)

	router := catTracks.NewRouter()

	http.Handle("/", router)

	fmt.Println("Listening on port ", porty)
	http.ListenAndServe(":"+strconv.Itoa(porty), nil)
}
