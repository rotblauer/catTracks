package catTracks

import (
	"github.com/rotblauer/tileTester2/note"
	"github.com/rotblauer/trackpoints/trackPoint"
	"testing"
	"time"
)

func TestValidVisitGrabbing(t *testing.T) {
	exampleNotesValidVisit := `{"floorsAscended":0,"customNote":"","currentTripStart":"2018-11-08T12:50:08.458Z","floorsDescended":0,"averageActivePace":3.100616996616512,"numberOfSteps":25,"visit":"{\"place\":\"25 Yeadon Ave, 25 Yeadon Ave, Charleston, SC  29407, United States @ <+32.78044829,-79.98285770> +\\\/- 100.00m, region CLCircularRegion (identifier:'<+32.78044828,-79.98285770> radius 141.76', center:<+32.78044828,-79.98285770>, radius:141.76m)\",\"arrivalDate\":\"2018-11-08T13:20:50.999Z\",\"validVisit\":true,\"departureDate\":\"4001-01-01T00:00:00.000Z\"}","relativeAltitude":-11.805885314941406,"currentCadence":0,"activity":"Stationary","currentPace":0,"pressure":101.96843719482422,"distance":19.160000000032596}`

	// Mon Jan 2 15:04:05 -0700 MST 2006
	// tt, err := time.Parse("2006-01-02T15:04:05", "2018-11-08T13:20:50.999Z")
	tt, err := time.Parse(time.RFC3339Nano, "2018-11-08T13:20:50.999Z")
	if err != nil {
		t.Fatal("err", err)
	}
	t.Log(tt)

	// Rye8 38.633697509765625 -90.26709747314453 0 0 -1 0 -1 0 2018-11-08 13:30:52.85 +0000 UTC {"floorsAs
	tp := trackPoint.TrackPoint{
		Uuid:      "9B4843BB-0EF7-4B54-832A-B6940304C531",
		ID:        1541683852850000000,
		Name:      "tester",
		Lat:       38.633697509765625,
		Lng:       -90.26709747314453,
		Accuracy:  0,
		Elevation: 0,
		Speed:     -1,
		Tilt:      0,
		Heading:   -1,
		HeartRate: 0,
		Time:      time.Now(),
		Notes:     exampleNotesValidVisit,
	}

	sn, err := note.NotesField(tp.Notes).AsNoteStructured()
	if err != nil {
		t.Error("err", err)
	}

	if !sn.HasValidVisit() {
		t.Error("invalid visit")
	}

	visit, err := sn.Visit.AsVisit()
	if err != nil {
		t.Error("err", err)
	}

	t.Log("visit", visit)
}
