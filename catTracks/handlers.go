package catTracks

//Handles
import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	// "html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/rotblauer/tileTester2/note"
	"github.com/rotblauer/trackpoints/trackPoint"
	"log"
	// "os"
	// "path"
)

// // the html stuff of this thing
// var templates = func() *template.Template {
// 	// p := path.Join(os.Getenv("GOPATH"), "src", "github.com", "rotblauer", "catTracks", "templates")
// 	// if _, err := os.Stat(p); err != nil {
// 	p := "templates"
// 	// }
// 	p = path.Join(p, "*.html")
// 	// p := path.Join(os.Getenv("GOPATH"), "src", "github.com", "rotblauer", "catTracks", "templates", "*.html")
// 	// p := path.Join(os.Getenv("GOPATH"), "src", "github.com", "rotblauer", "catTracks", "templates", "*.html")
// 	return template.Must(template.ParseGlob(p))
// }()

//W// elcome, loads and servers all (currently) data pointers
// func getIndexTemplate(w http.ResponseWriter, r *http.Request) {
// 	templates.ExecuteTemplate(w, "base", nil)
// }
// func getRaceTemplate(w http.ResponseWriter, r *http.Request) {
// 	templates.ExecuteTemplate(w, "race", nil)
// }
// func getMapTemplate(w http.ResponseWriter, r *http.Request) {
// 	templates.ExecuteTemplate(w, "map", nil)
// }
// func getLeafTemplate(w http.ResponseWriter, r *http.Request) {
// 	templates.ExecuteTemplate(w, "leaf", nil)
// }

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

var backlogPopulators [][]byte

// > https://stackoverflow.com/questions/24455147/how-do-i-send-a-json-string-in-a-post-request-in-go
// url := "http://restapi3.apiary.io/notes"
// fmt.Println("URL:>", url)

// var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
// req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
// req.Header.Set("X-Custom-Header", "myvalue")
// req.Header.Set("Content-Type", "application/json")

// client := &http.Client{}
// resp, err := client.Do(req)
// if err != nil {
// 	panic(err)
// }
// defer resp.Body.Close()

// fmt.Println("response Status:", resp.Status)
// fmt.Println("response Headers:", resp.Header)
// body, _ := ioutil.ReadAll(resp.Body)
// fmt.Println("response Body:", string(body))
func handleForwardPopulate(bod []byte) (err error) {

	if forwardPopulate == "" {
		log.Println("no forward url, not forwarding")
		return
	}

	backlogPopulators = append(backlogPopulators, bod)

	log.Println("forwarding to:", forwardPopulate, "#reqs:", len(backlogPopulators))

	var index int
	client := &http.Client{}

	for i, body := range backlogPopulators {
		index = i
		req, e := http.NewRequest("POST", forwardPopulate, bytes.NewBuffer(body))
		if e != nil {
			err = e
			break
		}
		req.Header.Set("Content-Type", "application/json")
		resp, e := client.Do(req)
		if e != nil {
			err = e
			break
		}
		err = resp.Body.Close()
		if err != nil {
			break
		}
	}

	if err == nil {
		backlogPopulators = [][]byte{}
	} else {
		backlogPopulators = backlogPopulators[index:]
	}

	return
}

var iftttWebhoook = "https://maker.ifttt.com/trigger/any_cat_visit/with/key/A_haNpM4rcpvsNYLFAy-8"

type IftttBodyCatVisit struct {
	Name    string `json:"value1"`
	Place   string `json:"value2"`
	MapsURL string `json:"value3"` // catonmap.net/{{value3}}
}

type IftttBodyCatVisit2 struct {
	Value1 string `json:"value1"`
	// April 29, 2013 at 12:01PM
	Value2 string `json:"value2"` // catonmap.net/{{value3}}
	Value3 int    `json:"value3"`
}

func populatePoints(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

	log.Println("handling pop:", r)

	var trackPoints trackPoint.TrackPoints

	var bod []byte
	var err error
	if forwardPopulate != "" {
		bod, err = ioutil.ReadAll(r.Body)
		// bod := []byte{}
		// n, err :=
		if err != nil {
			log.Println("err reading body", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// log.Println("read body ok, read nbytes=", len(bod))
		log.Println("read body ok, read nbytes=", len(bod))
		// log.Println("bod=", string(bod))
		// And now set a new body, which will simulate the same data we read:
		// > https://stackoverflow.com/questions/43021058/golang-read-request-body#43021236
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bod))
	}

	if r.Body == nil {
		log.Println("error: body nil")
		http.Error(w, "Please send a request body", 500)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&trackPoints)
	if err != nil {
		log.Println("error: decode json")
		http.Error(w, err.Error(), 400)
		return
	}

	// goroutine keeps http req from blocking while points are processed
	go func() {
		errS := storePoints(trackPoints)
		if errS != nil {
			log.Println("store err:", errS)
			// http.Error(w, errS.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("stored trackpoints", "len:", trackPoints.Len())
	}()

	// return empty json of empty trackpoints to not have to download tons of shit
	if errW := json.NewEncoder(w).Encode(&trackPoint.TrackPoints{}); errW != nil {
		log.Println("respond write err:", errW)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// if --forward-populate set, then make POST to set urls
	// --forward-populate=[]string{<downstream.urls.that.wants.points/put/em/here>}
	// goroutine keeps this request from block while pending this outgoing request
	// this keeps an original POST from being dependent on a forward POST
	go func() {
		if err := handleForwardPopulate(bod); err != nil {
			log.Println("forward populate error: ", err)
			// this just to persist any request that fails in case this process is terminated (backlogs are stored in mem)
			ioutil.WriteFile(fmt.Sprintf("dfp-%d", time.Now().UnixNano()), bod, 0666)
		} else {
			log.Println("forward populate finished OK")
		}
	}()

	// "visit":"{\"validVisit\":false}"
	for _, t := range trackPoints {
		ns, e := note.NotesField(t.Notes).AsNoteStructured()
		if e != nil {
			continue

		}
		if !ns.HasValidVisit() {
			continue
		}
		vis, err := ns.Visit.AsVisit()
		if err != nil {
			log.Println("error unmarshalling visit", err)
			continue
		}
		// info := IftttBodyCatVisit{
		// 	Name:  t.Name,
		// 	Place: vis.Place,
		// 	// x = lat, y = long
		// 	MapsURL: fmt.Sprintf("?z=%d&x=%.14f&y=%.14f&t=tile-dark&l=recent&s=", 14, t.Lat, t.Lng),
		// }

		place, err := vis.Place.AsPlace()
		if err != nil {
			log.Println("err parsing place", err)
			continue
		}

		mago := int(time.Now().Sub(vis.ArrivalTime).Round(time.Minute).Minutes())
		magoIFTTTword := fmt.Sprintf("%d minutes ago", mago)
		if mago == 1 {
			magoIFTTTword = fmt.Sprintf("%d minute ago", mago)
		} else if mago == 0 {
			magoIFTTTword = "Now"
		}

		info := IftttBodyCatVisit2{
			Value1: fmt.Sprintf(`%s visited %s

%s
%s

http://catonmap.net?z=%d&x=%.14f&y=%.14f&t=tile-dark&l=recent
`, t.Name, place.Identity, place.Identity, place.Address, 14, place.Lat, place.Lng),
			// April 29, 2013 at 12:01PM <-- ifttt
			// Mon Jan 02 15:04:05 -0700 2006 <-- go std templater
			// start date
			Value2: magoIFTTTword, // vis.ArrivalTime.Format("January _2, 2006") + " at " + vis.ArrivalTime.Format(time.Kitchen),
			Value3: int(vis.GetDuration().Round(time.Minute).Minutes()),
		}

		b, e := json.Marshal(info)
		if e != nil {
			log.Println("err marshal visit hooker", e)
			return
		}
		log.Println("sending ifttt webhook", "url=", iftttWebhoook, "info=", info)
		go func() {
			res, err := http.Post(iftttWebhoook, "application/json", bytes.NewBuffer(b))
			if err != nil {
				log.Println("err posting webhook", err)
				return
			}
			log.Println("webhook posted", "res", res.Status)
		}()
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
	w.Header().Add("Access-Control-Allow-Origin", "*")
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
	w.Header().Add("Access-Control-Allow-Origin", "*")

	b, e := getmetadata()
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got metadata:", len(b), "bytes")
	w.Write(b)
}
