package catTracks

// Handles saving and loading data
import (
	"github.com/rotblauer/trackpoints/trackPoint"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
)

// I think this has something to do with a table,,,,
var data = "TrackPoints"

//Store a snippit of life
func storePoint(trackPoint trackPoint.TrackPoint, c context.Context) error {
	key := trackPointKey(c)
	zeroTime := time.Time{}
	if trackPoint.Time == zeroTime {
		trackPoint.Time = time.Now()
	}
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
func getAllPoints(c context.Context, catQ string) []trackPoint.TrackPoint {
	q := datastore.NewQuery(data).Order("-Time")
	if catQ != "" {
		q = datastore.NewQuery(data).Filter("Name =", catQ).Order("-Time")
	}
	var ms []trackPoint.TrackPoint
	q.GetAll(c, &ms) //get em... this may be limited to 1000 though
	//log.Infof(c, "%#v", ms)
	return ms
}
