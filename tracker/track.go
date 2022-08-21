package tracker

import "time"

type Track struct {
	PosId int       `json:"positionId"`
	Loc   Loc       `json:"location"`
	Ts    time.Time `json:"timestamp"`
}
