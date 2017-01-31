package catTracks

import (
	"fmt"
	"math"
	"time"

	"github.com/deet/simpleline"
	"github.com/rotblauer/trackpoints/trackPoint"
)

// nope nope nope nope, hm. TODO.
// scale comes from ass/leaf-socket.js  getZoomLevel()
func getEpsFromScale(scale float64) float64 {
	n := math.Pow((scale / math.Pi), -1) // ~ 3.34916212
	fmt.Println("Scale", scale, " yields eps ", n)
	return n
}

func limitTrackPoints(query *query, tps trackPoint.TPs) (limitedTps trackPoint.TPs, e error) {

	start := time.Now()
	originalPointsCount := len(tps)

	var tpsSimple []simpleline.Point
	for _, tp := range tps {
		tpsSimple = append(tpsSimple, tp)
	}

	var epsilon = query.Epsilon
	// if query.Scale > 0.1 {
	// 	epsilon = getEpsFromScale(query.Scale)
	// }

	res, e := simpleline.RDP(tpsSimple, epsilon, simpleline.Euclidean, true)
	if e != nil {
		fmt.Println(e)
		return tps, e
	}

	for len(res) > query.Limit {
		epsilon = epsilon + epsilon/(1-epsilon)

		// could be rdp-ing the already rdp-ed?
		res2, e := simpleline.RDP(tpsSimple, epsilon, simpleline.Euclidean, true)
		if e != nil {
			fmt.Println("Error wiggling epsy.", e)
			res = tpsSimple
			continue
		} else {
			res = res2
		}
	}

	for _, simpleP := range res {
		tp, ok := simpleP.(*trackPoint.TrackPoint)
		if !ok {
			fmt.Println("shittt notok")
		}
		// could send channeler??
		limitedTps = append(limitedTps, tp)
	}

	fmt.Println("Limited ", originalPointsCount, " points to ", len(limitedTps), " in ", time.Since(start))
	return limitedTps, e
}
