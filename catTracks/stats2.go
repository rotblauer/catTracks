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

var debug = false

func debugLog(args ...interface{}) {
	if debug {
		log.Println(args...)
	}
}

type catStatsAggregate struct {
	Daily      catStatsCalculatedSlice
	Today      *catStatsCalculated // most recent
	ThreeDays  *catStatsCalculated
	SevenDays  *catStatsCalculated
	ThirtyDays *catStatsCalculated
	HalfYear   *catStatsCalculated
	Year       *catStatsCalculated
	AllTime    *catStatsCalculated
}

type catStatsCalculated struct {
	StartTime       time.Time
	Duration        time.Duration
	UserOrTeamStats []*userStats
}
type catStatsCalculatedSlice []*catStatsCalculated

func (s catStatsCalculatedSlice) Len() int {
	return len(s)
}

func (s catStatsCalculatedSlice) Less(i, j int) bool {
	// default sorter is 1,2,3,4,5
	return s[i].StartTime.After(s[j].StartTime) // because we want newest first
}

func (s catStatsCalculatedSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type userStats struct {
	Name  string // will also have "group" in addition to "Rye8" and "Big Papa"
	Raw   rawValues `json:"-"`
	Stats calcedMetrics
}

func (s *userStats) buildStatsFromRaw() *userStats {
	//debugLog(us.Name, len(us.Raw.Accuracy))

	us := &userStats{
		Name: s.Name,
		Raw:  s.Raw,
	}

	maxElevation, _ := us.Raw.Elevation.Max()
	maxSpeed, _ := us.Raw.Speed.Max()
	maxLat, _ := us.Raw.Lat.Max()
	maxLng, _ := us.Raw.Lng.Max()
	maxAccuracy, _ := us.Raw.Accuracy.Max()

	minElevation, _ := us.Raw.Elevation.Min()
	minSpeed, _ := us.Raw.Speed.Min()
	minLat, _ := us.Raw.Lat.Min()
	minLng, _ := us.Raw.Lng.Min()
	minAccuracy, _ := us.Raw.Accuracy.Min()

	medElevation, _ := us.Raw.Elevation.Median()
	medSpeed, _ := us.Raw.Speed.Median()
	medLat, _ := us.Raw.Lat.Median()
	medLng, _ := us.Raw.Lng.Median()
	medAccuracy, _ := us.Raw.Accuracy.Median()

	avgElevation, _ := us.Raw.Elevation.Mean()
	avgSpeed, _ := us.Raw.Speed.Mean()
	avgLat, _ := us.Raw.Lat.Mean()
	avgLng, _ := us.Raw.Lng.Mean()
	avgAccuracy, _ := us.Raw.Accuracy.Mean()

	stddevElevation, _ := us.Raw.Elevation.StandardDeviation()
	stddevSpeed, _ := us.Raw.Speed.StandardDeviation()
	stddevLat, _ := us.Raw.Lat.StandardDeviation()
	stddevLng, _ := us.Raw.Lng.StandardDeviation()
	stddevAccuracy, _ := us.Raw.Accuracy.StandardDeviation()

	varianceElevation, _ := us.Raw.Elevation.Variance()
	varianceSpeed, _ := us.Raw.Speed.Variance()
	varianceLat, _ := us.Raw.Lat.Variance()
	varianceLng, _ := us.Raw.Lng.Variance()
	varianceAccuracy, _ := us.Raw.Accuracy.Variance()

	sumElevation, _ := us.Raw.Elevation.Sum()
	sumSpeed, _ := us.Raw.Speed.Sum()
	sumLat, _ := us.Raw.Lat.Sum()
	sumLng, _ := us.Raw.Lng.Sum()
	sumAccuracy, _ := us.Raw.Accuracy.Sum()

	var absSumElevation float64
	for _, x := range us.Raw.Elevation {
		absSumElevation += math.Abs(x)
	}

	var absSumSpeed float64 // dumb
	for _, x := range us.Raw.Speed {
		absSumSpeed += math.Abs(x)
	}

	var absSumLat float64
	for _, x := range us.Raw.Lat {
		absSumLat += math.Abs(x)
	}

	var absSumLng float64
	for _, x := range us.Raw.Lng {
		absSumLng += math.Abs(x)
	}

	var absSumAccuracy float64 // dumb
	for _, x := range us.Raw.Accuracy {
		absSumAccuracy += math.Abs(x)
	}

	us.Stats.Elevation = calcedStats{
		Max:      maxElevation,
		Min:      minElevation,
		Avg:      avgElevation,
		Med:      medElevation,
		StdDev:   stddevElevation,
		Variance: varianceElevation,
		Sum:      sumElevation,
		AbsSum:   absSumElevation,
		Count: len(us.Raw.Elevation),
	}
	us.Stats.Speed = calcedStats{
		Max:      maxSpeed,
		Min:      minSpeed,
		Avg:      avgSpeed,
		Med:      medSpeed,
		StdDev:   stddevSpeed,
		Variance: varianceSpeed,
		Sum:      sumSpeed,
		AbsSum:   absSumSpeed,
		Count: len(us.Raw.Speed),
	}
	us.Stats.Lat = calcedStats{
		Max:      maxLat,
		Min:      minLat,
		Avg:      avgLat,
		Med:      medLat,
		StdDev:   stddevLat,
		Variance: varianceLat,
		Sum:      sumLat,
		AbsSum:   absSumLat,
		Count: len(us.Raw.Lat),
	}
	us.Stats.Lng = calcedStats{
		Max:      maxLng,
		Min:      minLng,
		Avg:      avgLng,
		Med:      medLng,
		StdDev:   stddevLng,
		Variance: varianceLng,
		Sum:      sumLng,
		AbsSum:   absSumLng,
		Count: len(us.Raw.Lng),
	}
	us.Stats.Accuracy = calcedStats{
		Max:      maxAccuracy,
		Min:      minAccuracy,
		Avg:      avgAccuracy,
		Med:      medAccuracy,
		StdDev:   stddevAccuracy,
		Variance: varianceAccuracy,
		Sum:      sumAccuracy,
		AbsSum:   absSumAccuracy,
		Count: len(us.Raw.Accuracy),
	}
	return us
}

type rawValues struct {
	Elevation Stats.Float64Data
	Speed     Stats.Float64Data
	Lat       Stats.Float64Data
	Lng       Stats.Float64Data
	Accuracy  Stats.Float64Data
}

type calcedMetrics struct {
	Elevation calcedStats
	Speed     calcedStats
	Lat       calcedStats
	Lng       calcedStats
	Accuracy  calcedStats
}
func (c rawValues) String() string {
	return fmt.Sprintf("acc=%.2f el=%.2f Speed=%.2f Lat=%.2f Lng=%.2f", c.Accuracy[0], c.Elevation[0], c.Speed[0], c.Lat[0], c.Lng[0])
}
func (c calcedMetrics) String() string {
	return fmt.Sprintf("acc=%.2f el=%.2f Speed=%.2f Lat=%.2f Lng=%.2f", c.Accuracy.Avg, c.Elevation.Avg, c.Speed.Avg, c.Lat.Avg, c.Lng.Avg)
}

type calcedStats struct {
	Max      float64
	Min      float64
	Avg      float64
	Med      float64
	StdDev   float64
	Variance float64
	Sum      float64 // Sum of (ironically absolute) values; How much Elevation did you traverse Today?
	AbsSum   float64 // Sum of values; How much higher or lower are you than when you started?
	Count    int
}

func (s *catStatsCalculated) getOrInitRawUserStats(name string) (*userStats, int) {
	for i, s := range s.UserOrTeamStats {
		if s.Name == name {
			return s, i
		}
	}
	return &userStats{Name: name}, -1
}

func (us *userStats) appendRawValues(point trackPoint.TrackPoint) {
	us.Raw.Elevation = append(us.Raw.Elevation, point.Elevation)
	us.Raw.Speed = append(us.Raw.Speed, point.Speed)
	us.Raw.Lat = append(us.Raw.Lat, point.Lat)
	us.Raw.Lng = append(us.Raw.Lng, point.Lng)
	us.Raw.Accuracy = append(us.Raw.Accuracy, point.Accuracy)
}

func (s *catStatsCalculated) createOrAppendRawValuesByUser(point trackPoint.TrackPoint) *catStatsCalculated {
	us, index := s.getOrInitRawUserStats(point.Name)
	us.appendRawValues(point)
	if index < 0 {
		s.UserOrTeamStats = append(s.UserOrTeamStats, us)
	} else {
		s.UserOrTeamStats[index] = us
	}
	return s
}

func (c *catStatsCalculatedSlice) getDaily(t time.Time) (int, *catStatsCalculated) {
	for i, d := range *c {
		if d.StartTime.Sub(t) < d.Duration {
			return i, d
		}
	}
	return -1, nil
}

func CalculateAndStoreStats(lastNDays int) error {
	// collect Raw values
	start := time.Now() // reference for dailies
	dailies := catStatsCalculatedSlice{
		&catStatsCalculated{
			StartTime: start,
			Duration:  24 * time.Hour,
		}}
	for i := 1; i <= lastNDays; i++ {
		dailies = append(dailies,
			&catStatsCalculated{
				StartTime: start.AddDate(0, 0, -i),
				Duration:  24 * time.Hour,
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
			// initialize new Daily batch
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

	debugLog("Raw", dailies[0].UserOrTeamStats[0].Name)
	debugLog("Raw", dailies[0].UserOrTeamStats[0].Raw)

	for i, d := range dailies {
		for j, s := range d.UserOrTeamStats {
			d.UserOrTeamStats[j] = s.buildStatsFromRaw()
			//debugLog(s)
		}
		dailies[i] = d
	}
	sort.Sort(dailies)

	debugLog("Stats", dailies[0].UserOrTeamStats[0].Name)
	debugLog("Stats", dailies[0].UserOrTeamStats[0].Stats)

	out := &catStatsAggregate{
		Daily: dailies,
		Today: dailies[0],
	}

	debugLog("agg_firstdaily", out.Daily[0].StartTime)
	debugLog("agg_firstdaily", out.Daily[0].UserOrTeamStats[0].Stats)
	debugLog("agg.Today", out.Today.UserOrTeamStats[0].Name)
	debugLog("agg.Today", out.Today.UserOrTeamStats[0].Stats)

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
			return errors.New("no data for Stats")
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
	fmt.Println("Got Stats:", len(b), "bytes")

	w.Write(b)
}
