package main

import (
	"flag"
	"github.com/rotblauer/catTracks/catTracks"
	"net/http"
	"strconv"
)

//start the url handlers, special init for everything?

//Toodle to do , Command line port arg, might mover er to main
func main() {
	var porty int
	flag.IntVar(&porty, "port", 8080, "port to serve and protect")
	flag.Parse()
	// Open Bolt DB.
	// catTracks.InitBoltDB()
	if bolterr := catTracks.InitBoltDB(); bolterr == nil {
		defer catTracks.GetDB().Close()
	}
	router := catTracks.NewRouter()
	//File server merveres
	ass := http.StripPrefix("/ass/", http.FileServer(http.Dir("./ass/")))
	router.PathPrefix("/ass/").Handler(ass)

	bower := http.StripPrefix("/bower_components/", http.FileServer(http.Dir("./bower_components/")))
	router.PathPrefix("/bower_components/").Handler(bower)

	http.Handle("/", router)

	http.ListenAndServe(":"+strconv.Itoa(porty), nil)
}
