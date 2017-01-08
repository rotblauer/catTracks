package catTracks

import "net/http"

//start the url handlers
func init() {
	router := NewRouter()
	http.Handle("/", router)
}
