package catTracks

//Handles
import (
	"encoding/csv"
	"encoding/json"
	"fmt"
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
	TrackPoints     []*trackPoint.TrackPoint
	TrackPointsJSON string
}

//Welcome, loads and servers all (currently) data pointers
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// catQ := r.FormValue("cat") //catQ is "" if not there //turn off queryable fur meow
	allPoints, e := getAllPoints()
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	pointsJSON, e := json.Marshal(allPoints)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	data := Data{TrackPoints: allPoints, TrackPointsJSON: string(pointsJSON)}
	fmt.Println("Done processing results")
	templates.Funcs(funcMap)
	templates.ExecuteTemplate(w, "base", data)
}

//TODO populate a population of points
func populatePoints(w http.ResponseWriter, r *http.Request) {
	var trackPoints trackPoint.TrackPoints

	if r.Body == nil {
		http.Error(w, "Please send a request body", 500)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&trackPoints)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	errS := storePoints(trackPoints)
	if errS != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//return json of trakcpoint if stored succcess
	if errW := json.NewEncoder(w).Encode(&trackPoints); errW != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadCSV(w http.ResponseWriter, r *http.Request) {
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

		errS := storePoint(tp)
		if errS != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	http.Redirect(w, r, "/", 302) //the 300

}
