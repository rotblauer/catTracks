/*
package main coerces RideWithGPS GPX data (for a ride) into CatTracks data,
printing the resulting GeoJSON FeatureCollection to stdout.
The RideWithGPS GPX data has a single LineString geometry, with a name, time, and coordTimes properties.
The LineString geometry is expected to have the same number of points as the coordTimes property.
The coordTimes property is expected to be an array of RFC3339 timestamps.
The GPX track points sometimes have 3 entries in their arrays, where the last entry is an elevation value.
The output is a GeoJSON FeatureCollection with a TrackPoint feature for each point in the LineString.
Novel CatTracks properties, like Name, UUID, and Activity are applied hard-coded in the TrackPoint struct.
*/
package main

import (
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/rotblauer/catTrackslib"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	// Read (all of) stdin and then parse it as a GeoJSON feature.
	read, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err)
	}

	fc, err := geojson.UnmarshalFeatureCollection(read)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Decoded %d features. BBOX bound: %v\n", len(fc.Features), fc.BBox.Bound())

	if len(fc.Features) != 1 {
		log.Fatalln("RideWithGPS exports are expected to have exactly one feature")
	}

	feature := fc.Features[0]

	if feature.Geometry.GeoJSONType() != "LineString" {
		log.Fatalln("RideWithGPS exports are expected to have a LineString geometry")
	}

	lineStringOrb := feature.Geometry.(orb.LineString)

	log.Printf("Decoded LineString with %d points.\n", len(lineStringOrb))

	expectedProperties := []string{"name", "time", "coordTimes"}
	for _, prop := range expectedProperties {
		if _, ok := feature.Properties[prop]; !ok {
			log.Fatalf("Missing expected property: %s\n", prop)
		} else {
			log.Printf("Found expected property: %s\n", prop)
		}
	}

	propName := feature.Properties.MustString("name")
	propTime := feature.Properties.MustString("time")
	lenCoordTimes := len(feature.Properties["coordTimes"].([]interface{}))

	log.Printf("Name: %s\n", propName)
	log.Printf("Time: %s\n", propTime)
	log.Printf("CoordTimes: %d\n", lenCoordTimes)

	if lenCoordTimes != len(lineStringOrb) {
		log.Fatalf("Mismatch between number of points in LineString (%d) and coordTimes (%d)\n", len(lineStringOrb), lenCoordTimes)
	}

	// Re-read the input as a JSON object, and extract the raw coordinates
	// in order to have easier access to the optional elevation entries.
	res := gjson.ParseBytes(read)
	rawCoords := res.Get("features.0.geometry.coordinates").Array()
	if len(rawCoords) != len(lineStringOrb) {
		log.Fatalf("Mismatch between number of points in LineString (%d) and raw coordinates (%d)\n", len(lineStringOrb), len(rawCoords))
	}

	output := geojson.NewFeatureCollection()

	for i, coord := range lineStringOrb {
		// Cross-reference and parse the time for the current point.
		coordTimeStr := feature.Properties["coordTimes"].([]interface{})[i].(string)
		trackTime, err := time.Parse(time.RFC3339, coordTimeStr)
		if err != nil {
			log.Fatalf("Error parsing time: %v\n", err)
		}

		// Elevation data is only included in some of the coordinates, in the third position.
		var elevation float64
		if len(rawCoords[i].Array()) == 3 {
			elevation = rawCoords[i].Array()[2].Float()
		}

		tp := catTrackslib.TrackPoint{
			Uuid:            "ia_elemnt_roam_2023",
			PushToken:       "thecattokenthatunlockstheworldtrackyourcats",
			Version:         "1",
			ID:              0,
			Name:            "Isaac's ELEMNT ROAM",
			Lat:             coord.Lat(),
			Lng:             coord.Lon(),
			Accuracy:        0,
			VAccuracy:       0,
			Elevation:       elevation,
			Speed:           -1,
			SpeedAccuracy:   0,
			Tilt:            0,
			Heading:         -1,
			HeadingAccuracy: -1,
			HeartRate:       -1,
			Time:            trackTime,
			Floor:           0,
			Notes:           `{"activity": "Bike"}`,
		}
		cattrackFeature := catTrackslib.TrackToFeature(&tp)
		output.Features = append(output.Features, cattrackFeature)
	}
	outputBytes, err := output.MarshalJSON()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(outputBytes))
}
