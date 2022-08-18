package tracker

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"time"

	"github.com/go-msvc/errors"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

type Grid interface {
	Add(*Vehicle) error
}

func Load(fn string, maxNr int, g Grid) error {
	//open the binary data file
	var totalSize int64
	if s, err := os.Stat(fn); err != nil {
		return errors.Wrapf(err, "cannot access %s", fn)
	} else {
		totalSize = s.Size()
	}
	log.Debugf("totalSize=%v", totalSize)

	f, err := os.Open(fn)
	if err != nil {
		return errors.Wrapf(err, "Cannot open file %s: %+v\n", fn, err)
	}
	defer f.Close()

	nrVehicles := 0
	t1 := time.Now()

	//all records are 30 bytes long...
	offset := 0
	for maxNr <= 0 || nrVehicles < maxNr {
		buffer := make([]byte, 30)
		nr, err := f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Debugf("failed to read: %+v\n", err)
			os.Exit(1)
		}
		if nr < 0 {
			log.Debugf("read returned %d\n", nr)
			os.Exit(1)
		}

		v := Vehicle{}
		v.PosId = int(binary.LittleEndian.Uint32(buffer[0:4])) //golang int = int64, which will be big enough and optimal for CPU
		v.Reg += string(buffer[4:14])
		v.Loc.Lat = float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[14:18]))) //use float64 on new CPU
		v.Loc.Lon = float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[18:22])))
		ts := binary.LittleEndian.Uint64(buffer[22:30])
		v.Ts = time.Unix(int64(ts), 0)
		if err := g.Add(&v); err != nil {
			log.Errorf("ofs:%d, buffer: %+v", offset, buffer)
			return errors.Wrapf(err, "offset %d, len %d: failed to add nr %d: %+v", offset, len(buffer), nrVehicles, v)
		}
		nrVehicles++
		totalSize -= 30
		if nrVehicles%100000 == 0 {
			log.Debugf("Read %10d: %+v, remain %v", nrVehicles, v, totalSize)
		}
	} //while reading
	t2 := time.Now()
	log.Debugf("t1-t2: %+v", t2.Sub(t1))
	return nil
} //Load()
