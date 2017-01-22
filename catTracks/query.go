package catTracks

import (
	"fmt"
	"github.com/creack/httpreq"
	"github.com/gorilla/mux"
	"net/http"
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

// Seems like non-trad form pass json is mux more mux friendly

type bounds struct {
	// NorthEast coords //etc
	// SouthWest coords
	NorthEastLat float64 `json:"northeastlat"`
	NorthEastLng float64 `json:"northeastlng"`
	SouthWestLat float64 `json:"southwestlat"`
	SouthWestLng float64 `json:"southwestlng"`
}

func parseQuery(r *http.Request, w http.ResponseWriter) *query {

	query := &query{}

	vars := mux.Vars(r)
	query.Version = vars["version"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}

	//Still not that great
	if err := httpreq.NewParsingMap().
		Add("isbounded", httpreq.ToBool, &query.IsBounded).
		Add("epsilon", httpreq.ToFloat64, &query.Epsilon).
		Add("northeastlast", httpreq.ToFloat64, &query.Bounds.NorthEastLat).
		Add("northeastlng", httpreq.ToFloat64, &query.Bounds.NorthEastLng).
		Add("southwestlat", httpreq.ToFloat64, &query.Bounds.SouthWestLat).
		Add("southwestlng", httpreq.ToFloat64, &query.Bounds.SouthWestLng).
		Parse(r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}

	fmt.Println("Processed query params as: ")
	fmt.Println("  API version: ", query.Version)
	fmt.Println("  Epsilon:     ", query.Epsilon)
	fmt.Println("  IsBounded:    ", query.IsBounded)
	fmt.Println("  Bounds:      ", query.Bounds)

	return query
}
