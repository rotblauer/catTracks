package catTracks

//Handles
import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"html/template"
	"net/http"
	"github.com/rotblauer/trackpoints/trackPoint"
)

var funcMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

type Data struct {
	TrackPoints []trackPoint.TrackPoint
}

//
//var tg = appengine.GeoPoint{Lat: 38.609896 + (rand.Float64() * 0.1), Lng: -90.331478 + (rand.Float64() * 0.1)}
//var test = trackPoint.TrackPoint{Elevation: 100.0, LatLong: tg, Time: time.Now()}
//
//storePoint(test, c)

//Welcome
func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	data := Data{TrackPoints: getAllPoints(c)}
	log.Infof(c, "Done processing results")
	templates.Funcs(funcMap)
	templates.ExecuteTemplate(w, "base", data)
}
