package catTracks

import (
	"fmt"
	"net/http"

	"github.com/creack/httpreq"
	"github.com/gorilla/mux"
	"github.com/rotblauer/trackpoints/trackPoint"
)

const (
	DefaultEpsilon = 0.001
	DefaultLimit   = 1000
)

type query struct {
	Epsilon float64 `json:"epsilon"`
	Version string  `json:"string"`
	Bounds  bounds  `json:"bounds"`
	Limit   int     `json:"limit"`
	Name    string  `json:"name"`
	Scale   float64 `json:"scale"`
}

// Seems like non-trad form pass json is mux more mux friendly
type bounds struct {
	NorthEastLat float64 `json:"northeastlat"`
	NorthEastLng float64 `json:"northeastlng"`
	SouthWestLat float64 `json:"southwestlat"`
	SouthWestLng float64 `json:"southwestlng"`
}

func (q *query) sumBounds() float64 {
	return q.Bounds.NorthEastLat + q.Bounds.NorthEastLng + q.Bounds.SouthWestLat + q.Bounds.SouthWestLng
}

func (q *query) IsBounded() bool {
	if q.sumBounds() == 0 || q.sumBounds() == 360.0*4.0 {
		return false
	}
	return true
}

func (q *query) PointInBounds(tp *trackPoint.TrackPoint) bool {
	var inY = tp.Lat < q.Bounds.NorthEastLat && tp.Lat > q.Bounds.SouthWestLat
	var inX = tp.Lng < q.Bounds.NorthEastLng && tp.Lng > q.Bounds.SouthWestLng
	return inY && inX
}

func (q *query) SetDefaults() {
	if q.Epsilon == 0 {
		q.Epsilon = DefaultEpsilon
	}
	if q.Limit == 0 {
		q.Limit = DefaultLimit
	}
}

func NewQuery() *query {
	return &query{
		Epsilon: DefaultEpsilon,
		Limit:   DefaultLimit,
	}
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
		// Add("isbounded", httpreq.ToBool, &query.IsBounded).
		Add("epsilon", httpreq.ToFloat64, &query.Epsilon).
		Add("northeastlat", httpreq.ToFloat64, &query.Bounds.NorthEastLat).
		Add("northeastlng", httpreq.ToFloat64, &query.Bounds.NorthEastLng).
		Add("southwestlat", httpreq.ToFloat64, &query.Bounds.SouthWestLat).
		Add("southwestlng", httpreq.ToFloat64, &query.Bounds.SouthWestLng).
		Add("limit", httpreq.ToInt, &query.Limit).
		Parse(r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}

	query.SetDefaults()

	fmt.Println("Processed query params as: ")
	fmt.Println("  API version: ", query.Version)
	fmt.Println("  Epsilon:     ", query.Epsilon)
	fmt.Println("  IsBounded:    ", query.IsBounded())
	fmt.Println("  Bounds:      ", query.Bounds)
	fmt.Println("  Limit:      ", query.Limit)

	return query
}
