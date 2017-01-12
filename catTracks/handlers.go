package catTracks

//Handles
import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"html/template"
	"net/http"
	"github.com/rotblauer/trackpoints/trackPoint"
	"encoding/json"
)

var funcMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

// the html stuff of this thing
var templates = template.Must(template.ParseGlob("templates/*.html"))

//For passing to the template
type Data struct {
	TrackPoints []trackPoint.TrackPoint
}

//Welcome, loads and servers all (currently) data pointers
func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	data := Data{TrackPoints: getAllPoints(c)}
	log.Infof(c, "Done processing results")
	templates.Funcs(funcMap)
	templates.ExecuteTemplate(w, "base", data)
}

func populatePoint(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var trackPoint trackPoint.TrackPoint

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&trackPoint)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	errS := storePoint(trackPoint, c)
	if errS != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}