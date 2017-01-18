package main

import (
	"net/http"
	"github.com/rotblauer/catTracks/catTracks"
)



//start the url handlers, special init for everything?

//Toodle to do , Command line port arg, might mover er to main
func main() {
	catTracks.InitBoltDB()
	router := catTracks.NewRouter()
	//File server merveres
	ass := http.StripPrefix("/ass/", http.FileServer(http.Dir("./ass/")))
	router.PathPrefix("/ass/").Handler(ass)

	bower := http.StripPrefix("/bower_components/", http.FileServer(http.Dir("./bower_components/")))
	router.PathPrefix("/bower_components/").Handler(bower)


	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}