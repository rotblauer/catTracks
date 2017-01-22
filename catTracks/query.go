package catTracks

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)


type query struct {
	Epsilon float64 `json:"epsilon"`
	Version string 	`json:"string"`
}

func SetDataAPI(router *mux.Router) {
	var h1 http.HandlerFunc
	h1 =getPointsJSON // I don't know why you must cast this

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
	fmt.Println("API version " + query.Version  +" with epsilon "+ query.Version)

	return query
}
