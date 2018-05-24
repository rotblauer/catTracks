package catTracks

//Handles
import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/rotblauer/trackpoints/trackPoint"
	"log"
)

// the html stuff of this thing
var templates = template.Must(template.ParseGlob("templates/*.html"))

//Welcome, loads and servers all (currently) data pointers
func getIndexTemplate(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "base", nil)
}
func getRaceTemplate(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "race", nil)
}
func getMapTemplate(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "map", nil)
}
func getLeafTemplate(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "leaf", nil)
}

func socket(w http.ResponseWriter, r *http.Request) {
	// see ./socket.go
	GetMelody().HandleRequest(w, r)
}

func getRaceJSON(w http.ResponseWriter, r *http.Request) {
	var e error

	var renderer = make(map[string]interface{})
	var spans = map[string]int{
		"today": 1,
		"week":  7,
		"all":   10,
	}

	for span, spanVal := range spans {
		renderer[span], e = buildTimePeriodStats(spanVal)
		if e != nil {
			fmt.Println(e)
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	}

	buf, e := json.Marshal(renderer)
	if e != nil {
		fmt.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	w.Write(buf)
}

func getPointsJSON(w http.ResponseWriter, r *http.Request) {
	query := parseQuery(r, w)

	data, eq := getData(query)
	if eq != nil {
		http.Error(w, eq.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Receive ajax get data string ")
	w.Write(data)
}
func getData(query *query) ([]byte, error) {
	var data []byte
	allPoints, e := getPointsQT(query)
	if e != nil {
		return data, e
	}
	data, err := json.Marshal(allPoints)
	if err != nil {
		return data, err
	}
	return data, nil
}

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
	go func() {
		errS := storePoints(trackPoints)
		if errS != nil {
			log.Println("store err:", errS)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Println("stored trackpoints", "len:", trackPoints.Len())
	}()

	// return empty json of empty trackpoints to not have to download tons of shit
	if errW := json.NewEncoder(w).Encode(&trackPoint.TrackPoints{}); errW != nil {
		log.Println("respond write err:", errW)
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

func getLastKnown(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin","*")
	b, e := getLastKnownData()
	//b, e := json.Marshal(lastKnownMap)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got lastknown:", len(b), "bytes")
	w.Write(b)
}

func getMetaData(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin","*")

	b, e := getmetadata()
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got metadata:", len(b), "bytes")
	w.Write(b)
}