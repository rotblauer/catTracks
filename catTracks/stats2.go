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
	"time"
	"encoding/binary"
	"strings"
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
	Name  string    // will also have "group" in addition to "Rye8" and "Big Papa"
	Raw   rawValues `json:"-"`
	Stats calcedMetrics
}

func sumDiff(ff Stats.Float64Data) float64 {
	var out float64
	for i, f := range ff {
		if i == 0 {
			continue
		}
		out += f - ff[i-1]
	}
	return out
}

func absSumDiff(ff Stats.Float64Data) float64 {
	var out float64
	for i, f := range ff {
		if i == 0 {
			continue
		}
		out += math.Abs(f - ff[i-1])
	}
	return out
}

func (s *userStats) buildStatsFromRaw() *userStats {
	//debugLog(us.Name, len(us.Raw.Accuracy))

	us := &userStats{
		Name: s.Name,
		Raw:  s.Raw,
	}

	for i, s := range us.Raw.Speed {
		if s < 0 {
			us.Raw.Speed[i] = 0
		}
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
		SumDiff: sumDiff(us.Raw.Elevation),
		AbsSumDiff: absSumDiff(us.Raw.Elevation),
		Count:    len(us.Raw.Elevation),
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
		SumDiff: sumDiff(us.Raw.Speed),
		AbsSumDiff: absSumDiff(us.Raw.Speed),
		Count:    len(us.Raw.Speed),
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
		SumDiff: sumDiff(us.Raw.Lat),
		AbsSumDiff: absSumDiff(us.Raw.Lat),
		Count:    len(us.Raw.Lat),
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
		SumDiff: sumDiff(us.Raw.Lng),
		AbsSumDiff: absSumDiff(us.Raw.Lng),
		Count:    len(us.Raw.Lng),
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
		SumDiff: sumDiff(us.Raw.Accuracy),
		AbsSumDiff: absSumDiff(us.Raw.Accuracy),
		Count:    len(us.Raw.Accuracy),
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
	Max        float64
	Min        float64
	Avg        float64
	Med        float64
	StdDev     float64
	Variance   float64
	Sum        float64
	AbsSum     float64
	SumDiff    float64 // Sum of (ironically absolute) values; How much Elevation did you traverse Today?
	AbsSumDiff float64 // Sum of values; How much higher or lower are you than when you started?
	Count      int
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

func (c *catStatsCalculatedSlice) getDaily(t time.Time) (int, *catStatsCalculated) {
	for i, d := range *c {
		if d.StartTime.Sub(t) < d.Duration {
			return i, d
		}
	}
	return -1, nil
}

// prefix = "storage": len = 7
//
var keyPrefixLen = len([]byte(statsDataKey))
var timeFmtLen = len([]byte(time.RFC3339Nano))
func (s *catStatsCalculated) buildStorageKey() []byte {
	var key []byte

	key = append(key, []byte(statsDataKey)...)
	key = s.StartTime.AppendFormat(key, time.RFC3339Nano)
	key = append(key, []byte("_")...)
	var b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(s.Duration.Nanoseconds()))
	key = append(key, b...)

	return key
}

func buildStorageKeyCheck(t time.Time, d time.Duration) []byte {
	var key []byte

	key = append(key, []byte(statsDataKey)...)
	key = t.AppendFormat(key, time.RFC3339Nano)
	key = append(key, []byte("_")...)
	var b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(d.Nanoseconds()))
	key = append(key, b...)

	return key
}

func getTimeAndSpanFromKey(key []byte) (time.Time, time.Duration, error) {
	//debugLog(string(key), len(key), keyLen, keyPrefixLen, keyPrefixLen+timeFmtLen)
	//if len(key) != keyLen {
	//	return time.Now(), 1*time.Second, fmt.Errorf("invalid key: %s", string(key))
	//}
	tbytes := key[keyPrefixLen:]
	s := strings.Split(string(tbytes), "_")
	
	//debugLog(string(s[0]), len(tbytes), tbytes)
	t, err := time.Parse(time.RFC3339Nano, s[0])
	if err != nil {
		return t, 0, err
	}

	dbytes := key[len(key)-timeFmtLen:]
	d := time.Duration(int64(binary.LittleEndian.Uint64(dbytes))) * time.Nanosecond
	return t, d, nil
}

func storeStats(s *catStatsCalculated) error {
	if s == nil {
		return errors.New("cannot store nil catStats")
	}
	val, e := json.Marshal(s)
	if e != nil {
		return e
	}
	if val == nil {
		return errors.New("no data to store")
	}

	debugLog("store: len(val)", len(val), "bytes")

	if e := GetDB().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		return b.Put(s.buildStorageKey(), val)
	}); e != nil {
		return e
	}
	return nil
}

// NOTE: spanStep should be +, spanOverall can be +/- where '-' means backward-looking and '+' means forward looking relative to time t
func CalculateAndStoreStatsByDateAndSpanStepping(t time.Time, spanStep, spanOverall time.Duration) error {
	tlim := t.Add(spanOverall)
	tPivot := t
	//  |t,tPivot --------------> |tlim
	if spanOverall > 0 {
		for tPivot.Before(tlim) {
			s, e := calculateStatsByDateAndSpan(tPivot, spanStep)
			if e != nil {
				return e
			}
			if e := storeStats(s); e != nil {
				return e
			}
			debugLog(">0", tPivot)
			tPivot = tPivot.Add(spanStep)
		}
	} else {
		tPivot = tlim
		// |tlim,tPivot ----------------> |t
		for tPivot.Before(t) {
			s, e := calculateStatsByDateAndSpan(tPivot, spanStep)
			if e != nil {
				return e
			}
			if e := storeStats(s); e != nil {
				return e
			}
			debugLog("<0", tPivot)
			tPivot = tPivot.Add(spanStep)
		}
	}
	return nil
}

func calculateStatsByDateAndSpan(t time.Time, span time.Duration) (*catStatsCalculated, error) {

	// check for pre-existence of immutable data
	var preExisting *catStatsCalculated
	if e := GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		v := b.Get(buildStorageKeyCheck(t, span))
		if v != nil {
			if err := json.Unmarshal(v, &preExisting); err != nil {
				return err
			}
		}
		return nil
	}); e != nil {
		return nil, e
	}
	if preExisting != nil {
		return preExisting, nil
	}

	// collect Raw values
	var daily *catStatsCalculated
	if e := GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(trackKey))
		e := b.ForEach(func(k, v []byte) error {
			var trackPointCurrent trackPoint.TrackPoint
			err := json.Unmarshal(v, &trackPointCurrent)
			if err != nil {
				return err
			}
			// break if beyond allow relative time frame
			if t.Sub(trackPointCurrent.Time) > span {
				return nil
			}
			if daily == nil {
				daily = &catStatsCalculated{
						StartTime: t,
						Duration: span,
					}
			}
			daily = daily.createOrAppendRawValuesByUser(trackPointCurrent)
			return nil
		})
		return e
	}); e != nil {
		return nil, e
	}
	if daily == nil {
		return nil, fmt.Errorf("no tracks in span: t=%v, d=%v", t, span)
	}

	for j, s := range daily.UserOrTeamStats {
		daily.UserOrTeamStats[j] = s.buildStatsFromRaw()
		//debugLog(s)
	}

	return daily, nil
}

// NOTE: use -duration to look backards, +duration to look forward relative to given time t
func getStatsByTimeSpan(t time.Time, d time.Duration) (catStatsCalculatedSlice, error) {
	var out catStatsCalculatedSlice
	startTime := t.Add(d)
	if e := GetDB().View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(statsKey))
		if e := b.ForEach(func(k, v []byte) error {
			debugLog(string(k), len(k))
			tk, td, e := getTimeAndSpanFromKey(k)
			debugLog("key", tk, td)
			if e != nil {
				debugLog("err:", e)
				return nil
			}
			// out of desired range
			if !(tk.Before(t) && tk.After(startTime)) {
				debugLog("stats out of bounds", "tk=", tk, "t=", t, "start=", startTime)
				return nil
			}
			var s = &catStatsCalculated{}
			if e := json.Unmarshal(v, s); e != nil {
				return e
			}
			out = append(out, s)
			return nil
		}); e != nil {
			return e
		}
		return nil
	}); e != nil {
		return nil, e
	}
	return out, nil
}

func GetStats(t time.Time, d time.Duration) ([]byte, error) {
	var ds catStatsAggregate
	s, err := getStatsByTimeSpan(t, d)
	if err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return nil, errors.New("empty dailies")
	}

	ds.Daily = s

	b, err := json.Marshal(ds)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func getStatsJSON(w http.ResponseWriter, r *http.Request) {
	data, e := GetStats(time.Now(), -24*time.Hour)
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
