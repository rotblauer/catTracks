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
	// Route{
	// 	"Index",
	// 	"GET",
	// 	"/",
	// 	getIndexTemplate,
	// },
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
		"WS",
		"GET",
		"/api/ws",
		socket,
	},
	// Route{
	// 	"Map",
	// 	"GET",
	// 	"/map",
	// 	getMapTemplate,
	// },
	// Route{
	// 	"Leaf",
	// 	"GET",
	// 	"/leaf",
	// 	getLeafTemplate,
	// },
	// Route{
	// 	"Race",
	// 	"GET",
	// 	"/race",
	// 	getRaceTemplate,
	// },
	Route{
		"RaceJSON",
		"GET",
		"/api/race",
		getRaceJSON,
	},
	Route{
		"StatsJSON",
		"GET",
		"/stats",
		getStatsJSON,
	},
	Route{
		"CatsLastKnown",
		"GET",
		"/lastknown",
		getLastKnown,
	},
	Route{
		"Metadata",
		"GET",
		"/metadata",
		getMetaData,
	},
	Route{
		"GetVisits",
		"GET",
		"/visits",
		handleGetPlaces,
	},
	Route{
		"GetVisitPhotos",
		"GET",
		"/googleNearbyPhotos",
		handleGetGoogleNearbyPhotos,
	},
}
