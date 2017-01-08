package catTracks

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func storePoint(trackPoint TrackPoint, c context.Context) error {
	key := trackPointKey(c)
	if _, err := datastore.Put(c, key, &trackPoint); err != nil { //store it
		return err
	}
	return nil
}

// forms the marker key
func trackPointKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, "TrackPoints", nil)
}
