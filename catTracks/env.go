package catTracks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rotblauer/trackpoints/trackPoint"
)

const (
	testesPrefix = "testes-------"
)

var testes = false
var forwardPopulate string

var tracksGZPath string
var tracksGZPathDevop string
var tracksGZPathEdge string

var masterdbpath string
var devopdbpath string
var edgedbpath string

var placesLayer bool

var (
	masterlock, devoplock, edgelock string
)

// SetTestes run
func SetTestes(flagOption bool) {
	testes = flagOption
}

// SetForwardPopulate sets the 'downstream' urls that should be forwarded
// any request that this client receives for populating points. Forward requests
// will be sent as POST requests in identical JSON as they are received.
// NOTE that forwardPopulate is a []string, so all uri's should be given in comma-separated
// format.
func SetForwardPopulate(arguments string) {
	forwardPopulate = arguments
	// // catch noop for legibility
	// if arguments == "" {
	// 	return
	// }
	// forwardPopulate = append(forwardPopulate, strings.Split(arguments, ",")...)
}

func SetLiveTracksGZ(pathto string) {
	tracksGZPath = pathto
}

func SetLiveTracksGZDevop(pathto string) {
	tracksGZPathDevop = pathto
}

func SetLiveTracksGZEdge(pathto string) {
	tracksGZPathEdge = pathto
}

func SetMasterLock(pathto string) {
	masterlock = pathto
}
func SetDevopLock(pathto string) {
	devoplock = pathto
}
func SetEdgeLock(pathto string) {
	edgelock = pathto
}

func SetDBPath(whichdb, pathto string) {
	switch whichdb {
	case "master", "":
		masterdbpath = pathto
	case "devop":
		devopdbpath = pathto
	case "edge":
		edgedbpath = pathto
	default:
		panic("invalid db name")
	}
}

func SetPlacesLayer(b bool) {
	placesLayer = b
}

func getTestesPrefix() string {
	if testes {
		return testesPrefix
	}
	return ""
}

// DeleteTestes wipes the entire database of all points with names prefixed with testes prefix. Saves an rm keystorke
func DeleteTestes() error {
	e := GetDB("master").Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var tp trackPoint.TrackPoint
			e := json.Unmarshal(v, &tp)
			if e != nil {
				fmt.Println("Error deleting testes.")
				return e
			}
			if strings.HasPrefix(tp.Name, testesPrefix) {
				b.Delete(k)
			}
		}
		return nil
	})
	return e
}
