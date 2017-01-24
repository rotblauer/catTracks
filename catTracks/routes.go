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
		"getPointsJSON",
		"GET",
		"/api/data/{version}",
		getPointsJSON,
	},
	Route{
		"Map",
		"GET",
		"/map",
		getMap,
	},
	Route{
		"Leaf",
		"GET",
		"/leaf",
		getLeaf,
	},
	Route{
		"Race",
		"GET",
		"/race",
		getRace,
	},
	Route{
		"RaceJSON",
		"GET",
		"/api/race",
		getRaceJSON,
	},
}
