package catTracks

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

var epsilonAPI = "epsilon"

type queryAPI struct {
	Epsilon float64 `json:"epsilon"`
}

func SetUpAPI(router *mux.Router) {
	var h1 http.HandlerFunc
	h1 =getPointsJSON // I don't know wh

	router.
	Methods("GET").
		Path("/{version}").
		Name("getPointsJSON").
		Handler(h1).Queries("epsilon", "{epsilon}")

}

func parseQuery(r *http.Request) queryAPI {
	var query queryAPI
	vars := mux.Vars(r)
	version := vars["version"]
	epsilon := vars["epsilon"]
	fmt.Println("API version " + version +" with epsilon "+epsilon)
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

	return query
}
