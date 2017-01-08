package catTracks

import "net/http"

//start the url handlers, special init for google app engine
func init() {
	router := NewRouter()
	http.Handle("/", router)
}
