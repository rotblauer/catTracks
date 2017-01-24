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

// the html stuff of this thing
var templates = template.Must(template.ParseGlob("templates/*.html"))

//Welcome, loads and servers all (currently) data pointers
func indexHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "base", nil)
}
func getRaceJSON(w http.ResponseWriter, r *http.Request) {
	var e error

	todayPoints, e := getPointsSince(time.Now().Add(-24 * time.Hour))
	if e != nil {
		fmt.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	weekPoints, e := getPointsSince(time.Now().Add(-24 * 7 * time.Hour))
	if e != nil {
		fmt.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	allPoints, e := getPointsSince(time.Now().Add(-24 * 5000 * time.Hour))
	if e != nil {
		fmt.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}

	//holy good damn this is ugly but i love it
	byCatToday := make(map[string]trackPoint.CatStats)
	byCatWeek := make(map[string]trackPoint.CatStats)
	byCatAll := make(map[string]trackPoint.CatStats)

	for _, name := range todayPoints.UniqueNames() { //erbody
		byCatToday[name] = todayPoints.ForName(name).Statistics()
	}
	for _, name := range weekPoints.UniqueNames() { //erbody
		byCatWeek[name] = weekPoints.ForName(name).Statistics()
	}
	for _, name := range allPoints.UniqueNames() { //erbody
		byCatAll[name] = allPoints.ForName(name).Statistics()
	}

	today := struct {
		TeamStats trackPoint.CatStats            `json:"team"`
		Cat       map[string]trackPoint.CatStats `json:"cat"`
	}{
		TeamStats: todayPoints.Statistics(),
		Cat:       byCatToday,
	}

	week := struct {
		TeamStats trackPoint.CatStats            `json:"team"`
		Cat       map[string]trackPoint.CatStats `json:"cat"`
	}{
		TeamStats: weekPoints.Statistics(),
		Cat:       byCatWeek,
	}

	all := struct {
		TeamStats trackPoint.CatStats            `json:"team"`
		Cat       map[string]trackPoint.CatStats `json:"cat"`
	}{
		TeamStats: weekPoints.Statistics(),
		Cat:       byCatAll,
	}

	var renderer = make(map[string]interface{})
	renderer["today"] = today
	renderer["week"] = week
	renderer["all"] = all

	// weekPoints, e := getPointsSince(time.Now().Add(-1 * time.Hour)) // could be better, slice off from todayPoints
	// if e != nil {
	// 	fmt.Println(e)
	// 	http.Error(w, e.Error(), http.StatusInternalServerError)
	// }

	// allPoints, e := getPointsSince(time.Now().AddDate(-100, 0, 0)) // also TODO
	// if e != nil {
	// 	fmt.Println(e)
	// 	http.Error(w, e.Error(), http.StatusInternalServerError)
	// }

	buf, e := json.Marshal(renderer)
	if e != nil {
		fmt.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	w.Write(buf)
}

func getRace(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "race", nil)
}

func getPointsJSON(w http.ResponseWriter, r *http.Request) {
	query := parseQuery(r, w)

	data, eq := getData(query)
	if eq != nil {
		http.Error(w, eq.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Receive ajax get data string ")
	w.Write([]byte(data))
}

func getMap(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "map", nil)
}
func getLeaf(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "leaf", nil)
}

func getData(query *query) ([]byte, error) {
	var data []byte
	allPoints, e := getAllPoints(query)
	if e != nil {
		return data, e
	}
	data, err := json.Marshal(allPoints)
	if err != nil {
		return data, err
	}
	return data, nil
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
