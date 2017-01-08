package catTracks

// Handles saving and loading data
import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	//"google.golang.org/appengine/log"
)

// I think this has something to do with a table,,,,
var data = "TrackPoints"

//Store a snippit of life
func storePoint(trackPoint TrackPoint, c context.Context) error {
	key := trackPointKey(c)
	if _, err := datastore.Put(c, key, &trackPoint); err != nil { //store it
		return err
	}
	return nil
}

// forms the incomplete key, think just i+1
func trackPointKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, data, nil)
}

//get everthing in the db... can do filtering some other day
func getAllPoints(c context.Context) []TrackPoint {
	q := datastore.NewQuery(data)
	var ms []TrackPoint
	q.GetAll(c, &ms)
	//log.Infof(c, "%#v", ms)
	return ms
}
