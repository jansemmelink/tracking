package tracker

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/go-msvc/errors"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelInfo)

//load vehicle tracks from file into memory
func Load(fn string) error {
	//open the binary data file
	f, err := os.Open(fn)
	if err != nil {
		return errors.Wrapf(err, "Cannot open file %s", fn)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return errors.Wrapf(err, "failed to read file %s", fn)
	}

	nrVehicles := 0
	t1 := time.Now()

	//all records are 30 bytes long...
	var minTs time.Time
	var maxTs time.Time
	var maxNrTracks int = 0

	maxNr := len(data) / 30
	for maxNr <= 0 || nrVehicles < maxNr {
		buffer := data[nrVehicles*30 : nrVehicles*30+30]
		var track Track
		track.PosId = int(binary.LittleEndian.Uint32(buffer[0:4])) //golang int = int64, which will be big enough and optimal for CPU
		reg := string(buffer[4:14])
		track.Loc.Lat = float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[14:18]))) //use float64 on new CPU
		track.Loc.Lon = float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[18:22])))
		ts := binary.LittleEndian.Uint64(buffer[22:30])
		track.Ts = time.Unix(int64(ts), 0)
		if v, ok := vehiclesByReg[reg]; ok {
			//append track to existing vehicles
			v.Tracks = append(v.Tracks, track)
			if len(v.Tracks) > maxNrTracks {
				maxNrTracks = len(v.Tracks)
			}
		} else {
			v := &Vehicle{
				Reg:    reg,
				Tracks: []Track{track},
			}
			vehiclesByReg[reg] = v
			if len(v.Tracks) > maxNrTracks {
				maxNrTracks = len(v.Tracks)
			}
		}
		if nrVehicles == 0 {
			minTs = track.Ts
			maxTs = track.Ts
		} else {
			if track.Ts.Before(minTs) {
				minTs = track.Ts
			}
			if track.Ts.After(maxTs) {
				maxTs = track.Ts
			}
		}
		nrVehicles++
		if nrVehicles%100000 == 0 {
			fmt.Printf("Loaded %d/%d\r", nrVehicles, maxNr)
		}
	} //while reading
	t2 := time.Now()
	log.Infof("Loaded %d vehicles tracks from %v..%v in %+v", nrVehicles, minTs, maxTs, t2.Sub(t1))
	log.Infof("Max tracks per vehicle is %d", maxNrTracks)
	return nil
} //Load()
