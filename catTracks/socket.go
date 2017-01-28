package catTracks

import (
	"encoding/json"
	"github.com/olahol/melody"
	// "github.com/rotblauer/trackpoints/trackPoint"
	"log"
)

var m *melody.Melody

//InitMelody does stuff
func InitMelody() {
	m = melody.New()

	// Incoming message about updated query params.
	m.HandleMessage(getPointsWS)
}

//GetMelody does stuff
func GetMelody() *melody.Melody {
	return m
}

// on request
func getPointsWS(s *melody.Session, msg []byte) {

	var q query
	log.Println("raw socket msg: ", string(msg))
	json.Unmarshal(msg, &q)
	log.Println("socket got query", q)

	// var c = make(chan *trackPoint.TrackPoint)
	pts, e := socketPointsByQueryQuadtree(&q)
	// pts, e := socketPointsByQueryGeohash(&q)
	if e != nil {
		log.Println("Couldn't get points.")
	}

	buf, e := json.Marshal(pts)
	if e != nil {
		log.Println("shit marshaling job by the socket")
	}

	s.Write(buf)

}

//on message

//on connect?
