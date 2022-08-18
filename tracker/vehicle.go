package tracker

import (
	"fmt"
	"math"
	"time"
)

type Loc struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (l1 Loc) Diff(l2 Loc) Loc {
	return Loc{
		Abs(l1.Lat - l2.Lat),
		Abs(l1.Lon - l2.Lon),
	}
}

func (l1 Loc) Dist(l2 Loc) float64 {
	dl := l1.Diff(l2)
	return math.Sqrt(dl.Lat*dl.Lat + dl.Lon*dl.Lon)
}

type Vehicle struct {
	PosId int       `json:"positionId"`
	Reg   string    `json:"registration"`
	Loc   Loc       `json:"location"`
	Ts    time.Time `json:"timestamp"`
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
	f.Write([]byte(fmt.Sprintf("V(%10.10s:%v)", v.Reg, v.Loc)))
}

func (l Loc) Format(f fmt.State, c rune) {
	f.Write([]byte(fmt.Sprintf("(%8.3f;%8.3f)", l.Lat, l.Lon)))
}
