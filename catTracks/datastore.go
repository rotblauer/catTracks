package catTracks

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	//"google.golang.org/appengine/log"
)

var data = "TrackPoints"

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

func getAllPoints(c context.Context) []TrackPoint {
	q := datastore.NewQuery(data)
	var ms []TrackPoint
	q.GetAll(c, &ms)
	//log.Infof(c, "%#v", ms)
	return ms
}
