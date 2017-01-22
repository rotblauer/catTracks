package catTracks

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type query struct {
	Epsilon   float64 `json:"epsilon"`
	Version   string  `json:"string"`
	Bounds    bounds  `json:"bounds"`
	IsBounded bool    `json:"isbounded"`
}

// iwould rather do pass json through the url. which can be done.
//ala
// type coords struct {
//    lat float64
//    lng float64
// }
type bounds struct {
	// NorthEast coords //etc
	// SouthWest coords
	NorthEastLat float64 `json:"northeastlat"`
	NorthEastLng float64 `json:"northeastlng"`
	SouthWestLat float64 `json:"southwestlat"`
	SouthWestLng float64 `json:"southwestlng"`
}

//SetDataAPI gets queries from request data
func SetDataAPI(router *mux.Router) {
	var h1 http.HandlerFunc
	h1 = getPointsJSON // I don't know why you must cast this

	router.
		Methods("GET").
		Path("/api/data/{version}").
		Name("getPointsJSON").
		Handler(h1).Queries("epsilon", "{epsilon}")
}

func parseQuery(r *http.Request) query {
	var query query
	vars := mux.Vars(r)
	query.Version = vars["version"] // not that anything ever changes

	epsilon := vars["epsilon"]

	if epsilon == "" {
		epsilon = "0.001"
	}
	eps, er := strconv.ParseFloat(epsilon, 64)
	if er != nil {
		fmt.Println("shit parsefloat eps")
		query.Epsilon = 0.001
	} else {
		query.Epsilon = eps
	}

	var isBounded = false
	fmt.Println("qisbound=", vars["isbounded"])
	fmt.Println("vars-", vars)
	if vars["isbounded"] != "" {
		isBounded = true
	}
	query.IsBounded = isBounded

	//i hoped it wouldn't come to this...
	if isBounded {
		bq := r.URL.Query()
		var nela, nelo, swla, swlo float64
		var e error
		var bounds bounds
		if qnela := bq.Get("northeastlast"); qnela != "" {
			nela, e = strconv.ParseFloat(qnela, 64)
			if e != nil {
				fmt.Println(e)
			}
		}
		if qnelo := bq.Get("northeastlng"); qnelo != "" {
			nelo, e = strconv.ParseFloat(qnelo, 64)
			if e != nil {
				fmt.Println(e)
			}
		}
		if qswla := bq.Get("southwestlat"); qswla != "" {
			swla, e = strconv.ParseFloat(qswla, 64)
			if e != nil {
				fmt.Println(e)
			}
		}
		if qswlo := bq.Get("southwestlng"); qswlo != "" {
			swlo, e = strconv.ParseFloat(qswlo, 64)
			if e != nil {
				fmt.Println(e)
			}
		}
		bounds.NorthEastLat = nela
		bounds.NorthEastLng = nelo
		bounds.SouthWestLat = swla
		bounds.SouthWestLng = swlo
		query.Bounds = bounds
	}

	fmt.Println("Processed query params as: ")
	fmt.Println("  API version: ", query.Version)
	fmt.Println("  Epsilon:     ", query.Epsilon)
	fmt.Println("  IsBounded:    ", query.IsBounded)
	fmt.Println("  Bounds:      ", query.Bounds)

	return query
}
