package catTracks

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/rotblauer/trackpoints/trackPoint"
)

const (
	testesPrefix = "testes-------"
)

var testes = false
var searchType = "quadtree"

// SetTestes run
func SetTestes(flagger bool) {
	testes = flagger
}

func getTestesPrefix() string {
	if testes {
		return testesPrefix
	}
	return ""
}

// SetSearch sets the env for kind of search to use
func SetSearch(s string) error {
	if s == "geohash" {
		searchType = "geohash"
		return nil
	}
	if s == searchType {
		return nil
	}
	return errors.New("Either 'quadtree' or 'geohash' for search type.")
}

// GetSearchType return the env search type
func GetSearchType() string {
	return searchType
}

// DeleteTestes wipes the entire database of all points with names prefixed with testes prefix. Saves an rm keystorke
func DeleteTestes() error {
	e := GetDB().Update(func(tx *bolt.Tx) error {
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
