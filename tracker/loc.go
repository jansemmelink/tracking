package tracker

import "math"

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
