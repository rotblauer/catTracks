package main

import (
	"flag"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/rotblauer/catTracks/catTracks"
)

//start the url handlers, special init for everything?

//Toodle to do , Command line port arg, might mover er to main
func main() {
	var porty int
	var clearDBTestes bool
	var testesRun bool
	var buildIndexes bool
	var forwardurl string
	var tracksjsongzpath string
	var tracksjsongzpathedge string
	var dbpath string
	var edgedbpath string

	flag.IntVar(&porty, "port", 8080, "port to serve and protect")
	flag.BoolVar(&clearDBTestes, "castrate-first", false, "clear out db of testes prefixed points") //TODO clear only certain values, ie prefixed with testes based on testesRun
	flag.BoolVar(&testesRun, "testes", false, "testes run prefixes name with testes-")              //hope that's your phone's name
	flag.BoolVar(&buildIndexes, "build-indexes", false, "build index buckets for original trackpoints")
	flag.StringVar(&forwardurl, "forward-url", "", "forward populate POST requests to this endpoint")
	flag.StringVar(&tracksjsongzpath, "tracks-gz-path", "", "path to appendable json.gz tracks (used by tippe)")
	flag.StringVar(&tracksjsongzpathedge, "edge-gz-path", "", "path to appendable json.gz tracks (used by tippe) - for edge tipping")
	flag.StringVar(&dbpath, "db-path-master", path.Join("db", "tracks.db"), "path to master tracks bolty db")
	flag.StringVar(&edgedbpath, "db-path-edge", path.Join("db", "edge.db"), "path to edge tracks bolty db")

	flag.Parse()

	catTracks.SetForwardPopulate(forwardurl)
	catTracks.SetLiveTracksGZ(tracksjsongzpath)
	catTracks.SetLiveTracksGZEdge(tracksjsongzpathedge)
	catTracks.SetDBPath("master", dbpath)
	catTracks.SetDBPath("edge", edgedbpath)

	// Open Bolt DB.
	// catTracks.InitBoltDB()
	if bolterr := catTracks.InitBoltDB(); bolterr == nil {
		defer catTracks.GetDB("master").Close()
	}
	if clearDBTestes {
		e := catTracks.DeleteTestes()
		if e != nil {
			log.Println(e)
		}
	}
	if buildIndexes {
		catTracks.BuildIndexBuckets() //cleverly always returns nil
	}
	// if qterr := catTracks.InitQT(); qterr != nil {
	// 	log.Println("Error initing QT.")
	// 	log.Println(qterr)
	// }
	catTracks.InitMelody()
	catTracks.SetTestes(testesRun) //is false defaulter, false prefixes names with ""

	router := catTracks.NewRouter()

	http.Handle("/", router)

	//go func() {
	//
	//}()

	//go func() {
	//	for {
	//		if e := catTracks.calculateAndStoreStats(7); e != nil {
	//			log.Println(e)
	//		}
	//	}
	//}()

	http.ListenAndServe(":"+strconv.Itoa(porty), nil)
}
