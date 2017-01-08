package catTracks

import (
	"google.golang.org/appengine"
	"html/template"
	"net/http"
)

var funcMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var tg = appengine.GeoPoint{Lat: 38.6270, Lng: -90.1994}
var test = TrackPoint{Elevation: 100.0, LatLong: tg}
var templates = template.Must(template.ParseGlob("templates/*.html"))

//Welcome
func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	storePoint(test, c)
	templates.Funcs(funcMap)

	templates.ExecuteTemplate(w, "base", nil)
}
