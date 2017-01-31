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
	m.HandleMessage(onMessageHandler)
}

//GetMelody does stuff
func GetMelody() *melody.Melody {
	return m
}

// on request
func onMessageHandler(s *melody.Session, msg []byte) {

	var q query
	json.Unmarshal(msg, &q)

	// var c = make(chan *trackPoint.TrackPoint)
	pts, e := getPointsWithQuery(&q)
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
