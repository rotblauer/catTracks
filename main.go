package main

import (
	"flag"
	"github.com/rotblauer/catTracks/catTracks"
	"log"
	"net/http"
	"strconv"
)

//start the url handlers, special init for everything?

//Toodle to do , Command line port arg, might mover er to main
func main() {
	var porty int
	var clearDBTestes bool
	var testesRun bool
	var noSpain bool

	flag.IntVar(&porty, "port", 8080, "port to serve and protect")
	flag.BoolVar(&clearDBTestes, "castrate-first", false, "clear out db of testes prefixed points") //TODO clear only certain values, ie prefixed with testes based on testesRun
	flag.BoolVar(&testesRun, "testes", false, "testes run prefixes name with testes-")              //hope that's your phone's name
	flag.BoolVar(&noSpain, "no-spain", false, "remove spanish wanderings")
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
	}
	if noSpain {
		e := catTracks.DeleteSpain()
		if e != nil {
			log.Println(e)
		}
	}
	catTracks.SetTestes(testesRun) //is false defaulter, false prefixes names with ""

	router := catTracks.NewRouter()
	//File server merveres
	ass := http.StripPrefix("/ass/", http.FileServer(http.Dir("./ass/")))
	router.PathPrefix("/ass/").Handler(ass)

	bower := http.StripPrefix("/bower_components/", http.FileServer(http.Dir("./bower_components/")))
	router.PathPrefix("/bower_components/").Handler(bower)

	catTracks.SetUpAPI(router)

	http.Handle("/", router)

	http.ListenAndServe(":"+strconv.Itoa(porty), nil)
}
