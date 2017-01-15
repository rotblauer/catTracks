package catTracks

//Handles
import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"encoding/csv"
	"encoding/json"
	"github.com/rotblauer/trackpoints/trackPoint"
	"html/template"
	"net/http"
	"strconv"
	"time"
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
	TrackPoints     []trackPoint.TrackPoint
	TrackPointsJSON string
}

//Welcome, loads and servers all (currently) data pointers
func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	catQ := r.FormValue("cat") //catQ is "" if not there
	log.Infof(c, "catQ: "+catQ)
	allPoints := getAllPoints(c, catQ)
	pointsJSON, e := json.Marshal(allPoints)
	if e != nil {
		log.Errorf(c, "Error making json from trackpoints.")
	}
	data := Data{TrackPoints: allPoints, TrackPointsJSON: string(pointsJSON)}
	log.Infof(c, "Done processing results")
	templates.Funcs(funcMap)
	templates.ExecuteTemplate(w, "base", data)
}

func populatePoint(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var trackPoint trackPoint.TrackPoint

	if r.Body == nil {
		http.Error(w, "Please send a request body", 500)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&trackPoint)
	if err != nil {
		log.Errorf(c, "Failed to decode: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}
	errS := storePoint(trackPoint, c)
	if errS != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//return json of trakcpoint if stored succcess
	if errW := json.NewEncoder(w).Encode(&trackPoint); errW != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadCSV(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	r.ParseMultipartForm(32 << 30)
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, line := range lines {
		var tp trackPoint.TrackPoint

		tp.Name = line[0]

		if tp.Time, err = time.Parse(time.UnixDate, line[1]); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tp.Lat, err = strconv.ParseFloat(line[2], 64); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tp.Lng, err = strconv.ParseFloat(line[3], 64); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		errS := storePoint(tp, c)
		if errS != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	http.Redirect(w, r, "/", 302) //the 300

}
