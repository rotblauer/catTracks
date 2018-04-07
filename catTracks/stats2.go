package catTracks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	Stats "github.com/montanaflynn/stats"
	"github.com/rotblauer/trackpoints/trackPoint"
	"log"
	"math"
	"net/http"
	"sort"
	"time"
)

var debug = true

func debugLog(args ...interface{}) {
	if debug {
		log.Println(args...)
	}
}

type catStatsAggregate struct {
	daily      catStatsCalculatedSlice
	today      *catStatsCalculated // most recent
	threeDays  *catStatsCalculated
	sevenDays  *catStatsCalculated
	thirtyDays *catStatsCalculated
	halfYear   *catStatsCalculated
	year       *catStatsCalculated
	allTime    *catStatsCalculated
}

type catStatsCalculated struct {
	startTime       time.Time
	duration        time.Duration
	userOrTeamStats []*userStats
}
type catStatsCalculatedSlice []*catStatsCalculated

func (s catStatsCalculatedSlice) Len() int {
	return len(s)
}

func (s catStatsCalculatedSlice) Less(i, j int) bool {
	// default sorter is 1,2,3,4,5
	return s[i].startTime.After(s[j].startTime) // because we want newest first
}

func (s catStatsCalculatedSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type userStats struct {
	name  string    // will also have "group" in addition to "Rye8" and "Big Papa"
	raw   rawValues `json:-`
	stats calcedMetrics
}

func (s *userStats) buildStatsFromRaw() *userStats {
	//debugLog(us.name, len(us.raw.accuracy))

	us := &userStats{
		name: s.name,
		raw:  s.raw,
	}

	maxElevation, _ := us.raw.elevation.Max()
	maxSpeed, _ := us.raw.speed.Max()
	maxLat, _ := us.raw.lat.Max()
	maxLng, _ := us.raw.lng.Max()
	maxAccuracy, _ := us.raw.accuracy.Max()

	minElevation, _ := us.raw.elevation.Min()
	minSpeed, _ := us.raw.speed.Min()
	minLat, _ := us.raw.lat.Min()
	minLng, _ := us.raw.lng.Min()
	minAccuracy, _ := us.raw.accuracy.Min()

	medElevation, _ := us.raw.elevation.Median()
	medSpeed, _ := us.raw.speed.Median()
	medLat, _ := us.raw.lat.Median()
	medLng, _ := us.raw.lng.Median()
	medAccuracy, _ := us.raw.accuracy.Median()

	avgElevation, _ := us.raw.elevation.Mean()
	avgSpeed, _ := us.raw.speed.Mean()
	avgLat, _ := us.raw.lat.Mean()
	avgLng, _ := us.raw.lng.Mean()
	avgAccuracy, _ := us.raw.accuracy.Mean()

	stddevElevation, _ := us.raw.elevation.StandardDeviation()
	stddevSpeed, _ := us.raw.speed.StandardDeviation()
	stddevLat, _ := us.raw.lat.StandardDeviation()
	stddevLng, _ := us.raw.lng.StandardDeviation()
	stddevAccuracy, _ := us.raw.accuracy.StandardDeviation()

	varianceElevation, _ := us.raw.elevation.Variance()
	varianceSpeed, _ := us.raw.speed.Variance()
	varianceLat, _ := us.raw.lat.Variance()
	varianceLng, _ := us.raw.lng.Variance()
	varianceAccuracy, _ := us.raw.accuracy.Variance()

	sumElevation, _ := us.raw.elevation.Sum()
	sumSpeed, _ := us.raw.speed.Sum()
	sumLat, _ := us.raw.lat.Sum()
	sumLng, _ := us.raw.lng.Sum()
	sumAccuracy, _ := us.raw.accuracy.Sum()

	var absSumElevation float64
	for _, x := range us.raw.elevation {
		absSumElevation += math.Abs(x)
	}

	var absSumSpeed float64 // dumb
	for _, x := range us.raw.speed {
		absSumSpeed += math.Abs(x)
	}

	var absSumLat float64
	for _, x := range us.raw.lat {
		absSumLat += math.Abs(x)
	}

	var absSumLng float64
	for _, x := range us.raw.lng {
		absSumLng += math.Abs(x)
	}

	var absSumAccuracy float64 // dumb
	for _, x := range us.raw.accuracy {
		absSumAccuracy += math.Abs(x)
	}

	us.stats.elevation = calcedStats{
		max:      maxElevation,
		min:      minElevation,
		avg:      avgElevation,
		med:      medElevation,
		stdDev:   stddevElevation,
		variance: varianceElevation,
		sum:      sumElevation,
		absSum:   absSumElevation,
	}
	us.stats.speed = calcedStats{
		max:      maxSpeed,
		min:      minSpeed,
		avg:      avgSpeed,
		med:      medSpeed,
		stdDev:   stddevSpeed,
		variance: varianceSpeed,
		sum:      sumSpeed,
		absSum:   absSumSpeed,
	}
	us.stats.lat = calcedStats{
		max:      maxLat,
		min:      minLat,
		avg:      avgLat,
		med:      medLat,
		stdDev:   stddevLat,
		variance: varianceLat,
		sum:      sumLat,
		absSum:   absSumLat,
	}
	us.stats.lng = calcedStats{
		max:      maxLng,
		min:      minLng,
		avg:      avgLng,
		med:      medLng,
		stdDev:   stddevLng,
		variance: varianceLng,
		sum:      sumLng,
		absSum:   absSumLng,
	}
	us.stats.accuracy = calcedStats{
		max:      maxAccuracy,
		min:      minAccuracy,
		avg:      avgAccuracy,
		med:      medAccuracy,
		stdDev:   stddevAccuracy,
		variance: varianceAccuracy,
		sum:      sumAccuracy,
		absSum:   absSumAccuracy,
	}
	return us
}

type rawValues struct {
	elevation Stats.Float64Data
	speed     Stats.Float64Data
	lat       Stats.Float64Data
	lng       Stats.Float64Data
	accuracy  Stats.Float64Data
}

type calcedMetrics struct {
	elevation calcedStats
	speed     calcedStats
	lat       calcedStats
	lng       calcedStats
	accuracy  calcedStats
}
func (c rawValues) String() string {
	return fmt.Sprintf("acc=%.2f el=%.2f speed=%.2f lat=%.2f lng=%.2f", c.accuracy[0], c.elevation[0], c.speed[0], c.lat[0], c.lng[0])
}
func (c calcedMetrics) String() string {
	return fmt.Sprintf("acc=%.2f el=%.2f speed=%.2f lat=%.2f lng=%.2f", c.accuracy.avg, c.elevation.avg, c.speed.avg, c.lat.avg, c.lng.avg)
}

type calcedStats struct {
	max      float64
	min      float64
	avg      float64
	med      float64
	stdDev   float64
	variance float64
	sum      float64 // sum of (ironically absolute) values; How much elevation did you traverse today?
	absSum   float64 // sum of values; How much higher or lower are you than when you started?
	count    int
}

func (s *catStatsCalculated) getOrInitRawUserStats(name string) (*userStats, int) {
	for i, s := range s.userOrTeamStats {
		if s.name == name {
			return s, i
		}
	}
	return &userStats{name: name}, -1
}

func (us *userStats) appendRawValues(point trackPoint.TrackPoint) {
	us.raw.elevation = append(us.raw.elevation, point.Elevation)
	us.raw.speed = append(us.raw.speed, point.Speed)
	us.raw.lat = append(us.raw.lat, point.Lat)
	us.raw.lng = append(us.raw.lng, point.Lng)
	us.raw.accuracy = append(us.raw.accuracy, point.Accuracy)
}

func (s *catStatsCalculated) createOrAppendRawValuesByUser(point trackPoint.TrackPoint) *catStatsCalculated {
	us, index := s.getOrInitRawUserStats(point.Name)
	us.appendRawValues(point)
	if index < 0 {
		s.userOrTeamStats = append(s.userOrTeamStats, us)
	} else {
		s.userOrTeamStats[index] = us
	}
	return s
}

func (c *catStatsCalculatedSlice) getDaily(t time.Time) (int, *catStatsCalculated) {
	for i, d := range *c {
		if d.startTime.Sub(t) < d.duration {
			return i, d
		}
	}
	return -1, nil
}

func CalculateAndStoreStats(lastNDays int) error {
	// collect raw values
	start := time.Now() // reference for dailies
	dailies := catStatsCalculatedSlice{
		&catStatsCalculated{
			startTime: start,
			duration:  24 * time.Hour,
		}}
	for i := 1; i <= lastNDays; i++ {
		dailies = append(dailies,
			&catStatsCalculated{
				startTime: start.AddDate(0, 0, -i),
				duration:  24 * time.Hour,
			})
	}

	if e := GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))
		e := b.ForEach(func(k, v []byte) error {
			var trackPointCurrent trackPoint.TrackPoint
			err := json.Unmarshal(v, &trackPointCurrent)
			if err != nil {
				return err
			}
			// break if beyond allow relative time frame
			if start.Sub(trackPointCurrent.Time) > time.Duration(lastNDays)*24*time.Hour {
				return nil
			}
			// initialize new daily batch
			index, stat := dailies.getDaily(trackPointCurrent.Time)
			if stat == nil {
				return nil
			}
			dailies[index] = stat.createOrAppendRawValuesByUser(trackPointCurrent)
			return nil
		})
		return e
	}); e != nil {
		return e
	}

	debugLog("raw", dailies[0].userOrTeamStats[0].name)
	debugLog("raw", dailies[0].userOrTeamStats[0].raw)

	for i, d := range dailies {
		for j, s := range d.userOrTeamStats {
			d.userOrTeamStats[j] = s.buildStatsFromRaw()
			//debugLog(s)
		}
		dailies[i] = d
	}
	sort.Sort(dailies)

	debugLog("stats", dailies[0].userOrTeamStats[0].name)
	debugLog("stats", dailies[0].userOrTeamStats[0].stats)

	out := &catStatsAggregate{
		daily: dailies,
		today: dailies[0],
	}

	debugLog("agg_firstdaily", out.daily[0].startTime)
	debugLog("agg_firstdaily", out.daily[0].userOrTeamStats[0].stats)
	debugLog("agg.today", out.today.userOrTeamStats[0].name)
	debugLog("agg.today", out.today.userOrTeamStats[0].stats)

	val, e := json.Marshal(out)
	if e != nil {
		return e
	}
	if val == nil {
		return errors.New("no data to store")
	}

	debugLog("len(val)", len(val))

	if e := GetDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		return b.Put([]byte(statsDataKey), val)
	}); e != nil {
		return e
	}
	return nil
}

func GetStats() ([]byte, error) {
	var out []byte
	if e := GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		val := b.Get([]byte(statsDataKey))
		if val == nil {
			return errors.New("no data for stats")
		}
		out = val
		return nil
	}); e != nil {
		return nil, e
	}
	return out, nil
}

func getStatsJSON(w http.ResponseWriter, r *http.Request) {
	data, e := GetStats()
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}

	b, e := json.Marshal(data)
	if e != nil {
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
	fmt.Println("Got stats:", len(b), "bytes")

	w.Write(b)
}
