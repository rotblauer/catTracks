package catTracks

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type query struct {
	Epsilon float64 `json:"epsilon"`
	Version string  `json:"string"`
	Bounds  bounds  `json:"bounds"`
}
type bounds struct {
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

	//i hoped it wouldn't come to this...
	bq := r.URL.Query()
	var bounds bounds
	nela, e := strconv.ParseFloat(bq.Get("northeastlat"), 64)
	if e != nil {
		fmt.Println(e)
	}
	nelo, e := strconv.ParseFloat(bq.Get("northeastlng"), 64)
	if e != nil {
		fmt.Println(e)
	}
	swla, e := strconv.ParseFloat(bq.Get("southwestlat"), 64)
	if e != nil {
		fmt.Println(e)
	}
	swlo, e := strconv.ParseFloat(bq.Get("southwestlng"), 64)
	if e != nil {
		fmt.Println(e)
	}
	bounds.NorthEastLat = nela
	bounds.NorthEastLng = nelo
	bounds.SouthWestLat = swla
	bounds.SouthWestLng = swlo

	fmt.Println("API version "+query.Version+" with epsilon ", query.Epsilon)
	fmt.Println("Bounds: ", bounds)

	return query
}
