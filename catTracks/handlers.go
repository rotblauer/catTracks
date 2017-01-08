package catTracks
//Handles
import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"html/template"
	"net/http"
	"time"
	"math/rand"

)

var funcMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var tg = appengine.GeoPoint{Lat: 38.6270+(rand.Float64()*3.0), Lng: -90.1994+(rand.Float64()*3.0)}
var test = TrackPoint{Elevation: 100.0, LatLong: tg, Time: time.Now()}
var templates = template.Must(template.ParseGlob("templates/*.html"))

type Data struct {
	TrackPoints []TrackPoint
}

//Welcome
func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	storePoint(test, c)
	data := Data{TrackPoints: getAllPoints(c)}
	log.Infof(c, "Done processing results")

	templates.Funcs(funcMap)
	templates.ExecuteTemplate(w, "base", data)
}
