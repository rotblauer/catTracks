package catTracks

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	// "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/schema"

	"github.com/rotblauer/tileTester2/note"
	"github.com/rotblauer/trackpoints/trackPoint"
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

// W// elcome, loads and servers all (currently) data pointers
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

type forwardingQueueItem struct {
	payload []byte
	request *http.Request
}

var backlogPopulators []*forwardingQueueItem

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
func handleForwardPopulate(r *http.Request, bod []byte) (err error) {

	if forwardPopulate == "" {
		log.Println("no forward url, not forwarding")
		return
	}

	backlogPopulators = append(backlogPopulators, &forwardingQueueItem{
		request: r,
		payload: bod,
	})

	log.Println("forwarding to:", forwardPopulate, "#reqs:", len(backlogPopulators))

	var index int
	client := &http.Client{}

	for i, fqi := range backlogPopulators {
		index = i
		req, e := http.NewRequest("POST", forwardPopulate, bytes.NewBuffer(fqi.payload))
		if e != nil {
			err = e
			break
		}

		// type Header map[string][]string
		for k, v := range fqi.request.Header {
			for _, vv := range v {
				req.Header.Set(k, vv)
			}
		}

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
		backlogPopulators = []*forwardingQueueItem{}
	} else {
		if index < len(backlogPopulators) {
			backlogPopulators = append(backlogPopulators[:index], backlogPopulators[index+1:]...)
		} else {
			backlogPopulators = []*forwardingQueueItem{}
		}

		log.Println("forwarding error:", err, "index", index, "len backlog", len(backlogPopulators))
	}

	return
}

var iftttWebhoook = "https://maker.ifttt.com/trigger/any_cat_visit/with/key/" + os.Getenv("IFTTT_WEBHOOK_TOKEN")

type IftttBodyCatVisit struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
	Value3 int    `json:"value3"`
}

// ToJSONbuffer converts some newline-delimited JSON to valid JSON buffer
func toJSONbuffer(reader io.Reader) []byte {
	// var buffer bytes.Buffer

	// buffer.Write([]byte("["))

	reg := regexp.MustCompile(`(?m)\S*`)
	out := []byte("[")
	scanner := bufio.NewScanner(reader)
	for {
		ok := scanner.Scan()
		if ok {
			sb := scanner.Bytes()
			if reg.Match(sb) {
				out = append(out, scanner.Bytes()...)
				out = append(out, []byte(",")...)
			}
			continue
		}
		break
	}
	out = bytes.TrimSuffix(out, []byte(","))
	out = append(out, []byte{byte(']'), byte('\n')}...)

	// r := bufio.NewReader(reader)
	//
	// buffer.Write([]byte("["))
	// for {
	//	bytes, err := r.ReadBytes(byte('\n'))
	//	//bytes, _, err := r.ReadLine()
	//	buffer.Write(bytes)
	//	//r.Peek(1)
	//	if err == io.EOF || string(bytes) == "" {
	//		break
	//	}
	//	buffer.Write([]byte(","))
	// }
	//
	// bu := []byte{}
	// buffer.Write(bu)
	// bu = bytes.TrimSuffix(bu, []byte(","))
	//
	// buffer.Reset()
	// buffer.Write(bu)
	//
	// //if bytes.Equal(buffer.Bytes()[buffer.Len()-1:], []byte(",")) {
	// //	buffer.UnreadByte()
	// //}
	//
	// buffer.Write([]byte("]"))
	// buffer.Write([]byte("\n"))

	return out
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

	bod, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error reading body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(bod, &trackPoints)
	// err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(bod))).Decode(&trackPoints)
	if err != nil {
		log.Println("Could not decode json as array, body length was:", len(bod))

		// try decoding as ndjson..
		ndbod := toJSONbuffer(ioutil.NopCloser(bytes.NewBuffer(bod)))

		log.Println("attempting decode as ndjson instead..., length:", len(ndbod), string(ndbod))

		// err = json.NewDecoder(&ndbuf).Decode(&trackPoints)
		err = json.Unmarshal(ndbod, &trackPoints)
		if err != nil {
			log.Println("could not decode req as ndjson, error:", err.Error())

			// err = json.Unmarshal(json.RawMessage(bod), &trackPoints)

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			log.Println("OK: decoded request as ndjson instead")
		}
	}

	log.Println("checking token")
	tok := os.Getenv("COTOKEN")
	if tok == "" {
		log.Println("ERROR: no COTOKEN env var set")
	} else {
		log.Println("GOODNEWS: using COTOKEN for cat verification")
		log.Println()
		if b, _ := httputil.DumpRequest(r, true); b != nil {
			log.Println(string(b))
		}
		log.Println()
		verified := false
		headerKey := "AuthorizationOfCats"
		if h := r.Header.Get(headerKey); h != "" {
			log.Println("using header verification...")
			if h == tok {
				log.Println("header OK")
				verified = true
			} else {
				log.Println("header verification failed: ", h)
			}
		} else {
			// catonmap.info:3001/populate?api_token=asdfasdfb
			r.ParseForm()
			if token := r.FormValue("api_token"); token != "" {
				if token == tok {
					log.Println("used token verification: OK")
					verified = true
				} else {
					log.Println("token verification failed:", token)
					verified = true
				}
			}
		}
		if verified {
			trackPoints.Verified()
			log.Println("GOODNEWS: verified cattracks posted remote.host:", r.RemoteAddr)
		} else {
			trackPoints.Unverified(r)
			log.Println("WARNING: unverified cattracks posted remote.host:", r.RemoteAddr)
		}
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
		if err := handleForwardPopulate(r, bod); err != nil {
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

		place, err := vis.Place.AsPlace()
		if err != nil {
			log.Println("error parsing place", err)
			continue
		}

		mago := int(time.Now().Sub(vis.ArrivalTime).Round(time.Minute).Minutes())
		magoIFTTTword := fmt.Sprintf("%d minutes ago", mago)
		if mago == 1 {
			magoIFTTTword = fmt.Sprintf("%d minute ago", mago)
		} else if mago == 0 {
			magoIFTTTword = "Now"
		}

		info := IftttBodyCatVisit{
			Value1: fmt.Sprintf(`%s visited %s

%s
%s

http://catonmap.net?z=%d&x=%.14f&y=%.14f&t=tile-dark&l=recent
`, t.Name, place.Identity, place.Identity, place.Address, 14, place.Lat, place.Lng),
			// April 29, 2013 at 12:01PM <-- ifttt (output fmt), unknown fmt for input
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
		// go func() {
		// 	catmapurl := fmt.Sprintf("http://catonmap.net?z=%d&x=%.14f&y=%.14f", 14, place.Lat, place.Lng)
		// 	p := struct {
		// 		Value1 string `json:"value1"`
		// 		Value2 string `json:"value2"`
		// 		Value3 string `json:"value3"`
		// 	}{
		// 		Value1: t.Name,
		// 		Value2: place.Identity,
		// 		// Value3:
		// 	}
		// 	b, e := json.Marshal(b)
		// 	if e != nil {
		// 		log.Println("err marshalling isaac cat track hook", err)
		// 		return
		// 	}
		// 	url := strings.Replace(iftttWebhoook, "any_cat_visit", "cat_visit_ia_twitter", -1)
		// 	log.Println("sending ifttt webhook @isaac", "url=", url, "info=", p)
		// 	res, err := http.Post(url, "application/json", bytes.NewBuffer(b))
		// 	if err != nil {
		// 		log.Println("err posting webhook", err)
		// 		return
		// 	}
		// 	log.Println("webhook posted", "res", res.Status)
		// }()
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
		var tp *trackPoint.TrackPoint

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

		_, errS := storePoint(tp)
		if errS != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	http.Redirect(w, r, "/", 302) // the 300

}

func getLastKnown(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	b, e := getLastKnownData()
	// b, e := json.Marshal(lastKnownMap)
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

var decoder = schema.NewDecoder()

// returns response type image
func handleGetGoogleNearbyPhotos(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
	}
	var qf QueryFilterGoogleNearbyPhotos
	err = decoder.Decode(&qf, r.Form) // note using r.Form, not r.PostForm
	if err != nil {
		http.Error(w, "err decoding request: "+err.Error(), http.StatusBadRequest)
	}

	b, e := getGoogleNearbyPhotos(qf)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got googlenearby photos:", len(b), "bytes")
	w.Write(b)
}

func handleGetPlaces(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// parse params
	// NOTE:
	// func (r *Request) ParseForm() error
	// ParseForm populates r.Form and r.PostForm.
	//
	// For all requests, ParseForm parses the raw query from the URL and
	// updates r.Form.

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
	}
	var qf QueryFilterPlaces
	err = decoder.Decode(&qf, r.Form) // note using r.Form, not r.PostForm
	if err != nil {
		http.Error(w, "err decoding request: "+err.Error(), http.StatusBadRequest)
	}

	b, e := getPlaces(qf)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got places:", len(b), "bytes")
	w.Write(b)
}

func handleGetPlaces2(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// parse params
	// NOTE:
	// func (r *Request) ParseForm() error
	// ParseForm populates r.Form and r.PostForm.
	//
	// For all requests, ParseForm parses the raw query from the URL and
	// updates r.Form.

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
	}
	var qf QueryFilterPlaces
	err = decoder.Decode(&qf, r.Form) // note using r.Form, not r.PostForm
	if err != nil {
		http.Error(w, "err decoding request: "+err.Error(), http.StatusBadRequest)
	}

	b, e := getPlaces2(qf)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got places:", len(b), "bytes")
	w.Write(b)
}

func handleGetCatSnaps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	var startQ, endQ time.Time
	startRaw, ok := r.URL.Query()["tstart"]
	if ok && len(startRaw) > 0 {
		i64, err := strconv.ParseInt(startRaw[0], 10, 64)
		if err == nil {
			startQ = time.Unix(i64, 0)
		} else {
			log.Printf("catsnaps: Invalid t-start value: %s (%v)\n", startRaw[0], err)
		}
	}
	endRaw, ok := r.URL.Query()["tend"]
	if ok && len(endRaw) > 0 {
		i64, err := strconv.ParseInt(endRaw[0], 10, 64)
		if err == nil {
			endQ = time.Unix(i64, 0)
		} else {
			log.Printf("catsnaps: Invalid t-end value: %s (%v)\n", endRaw[0], err)
		}
	}
	b, e := getCatSnaps(startQ, endQ)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got catsnaps", len(b), "bytes")
	w.Write(b)
}
