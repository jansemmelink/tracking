package tracker

import (
	"fmt"
	"time"

	"github.com/go-msvc/errors"
)

var vehiclesByReg = map[string]*Vehicle{}

func NrVehicles() int {
	return len(vehiclesByReg)
}

type Vehicle struct {
	Reg    string  `json:"registration"`
	Tracks []Track `json:"track"`
}

//get vehicle location as specified time, interpolating as necessary
func (v Vehicle) LocAtTs(ts time.Time) (Loc, error) {
	if len(v.Tracks) == 0 {
		return Loc{}, errors.Errorf("no tracks")
	}
	n := len(v.Tracks)
	if ts.After(v.Tracks[n-1].Ts) {
		return v.Tracks[n-1].Loc, nil
	}
	if ts.Before(v.Tracks[0].Ts) {
		return v.Tracks[0].Loc, nil
	}
	//find the track with closest ts
	min := 0
	max := n - 1
	for min != max {
		i := (min + max) / 2
		if v.Tracks[i].Ts.After(ts) {
			max = i - 1
			continue
		}
		if v.Tracks[i].Ts.Before(ts) {
			min = 1 + 1
			continue
		}
		min = i
		max = i
		break
	}
	if min == max {
		return v.Tracks[min].Loc, nil
	}
	//interpolate
	t1 := v.Tracks[min]
	t2 := v.Tracks[max]
	td := t1.Ts.Sub(t2.Ts)
	dLat := t2.Loc.Lat - t1.Loc.Lat
	dLon := t2.Loc.Lon - t1.Loc.Lon
	tx := ts.Sub(t1.Ts)
	loc := Loc{
		t1.Loc.Lat + dLat*float64(tx/td),
		t1.Loc.Lon + dLon*float64(tx/td),
	}
	log.Debugf("Int %s: %+v .. %+v -> @ %v: %+v",
		v.Reg,
		t1,
		t2,
		ts,
		loc)
	return loc, nil
}

func Abs(f1 float64) float64 {
	if f1 < 0 {
		return -f1
	}
	return f1
}

func MaxOf(f1, f2 float64) float64 {
	if f1 > f2 {
		return f1
	}
	return f2
}

func (v Vehicle) Format(f fmt.State, c rune) {
	f.Write([]byte(fmt.Sprintf("V(%10.10s:%v)", v.Reg, v.Tracks[len(v.Tracks)-1].Loc)))
}

func (l Loc) Format(f fmt.State, c rune) {
	f.Write([]byte(fmt.Sprintf("(%8.3f;%8.3f)", l.Lat, l.Lon)))
}
