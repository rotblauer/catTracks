package catTracks

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		indexHandler,
	},
	Route{
		"PointPopulator",
		"POST",
		"/populate/",
		populatePoints,
	},
	Route{
		"UploadCSV",
		"POST",
		"/upload",
		uploadCSV,
	},
	Route{
		"Map",
		"GET",
		"/map",
		getMap,
	},
	Route{
		"GetPointsJSON",
		"GET",
		"/v1",
		getPointsJSON,
	},
	Route{
		"Leaf",
		"GET",
		"/leaf",
		getLeaf,
	},
}
