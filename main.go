package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jansemmelink/tracking/tracker"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func main() {
	dataFilename := flag.String("input", "./data/VehiclePositions.dat", "Binary input file")
	searchFilename := flag.String("search", "./search1.json", "File with list of locations to search")
	flag.Parse()

	//read the locations to search
	f, err := os.Open(*searchFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open --search=%s: %+v\n", *searchFilename, err)
		os.Exit(1)
	}
	searches := []Search{}
	if err := json.NewDecoder(f).Decode(&searches); err != nil {
		f.Close()
		fmt.Fprintf(os.Stderr, "Cannot read JSON from --search=%s: %+v\n", *searchFilename, err)
		os.Exit(1)
	}
	f.Close()
	f = nil

	if len(searches) == 0 {
		fmt.Fprintf(os.Stderr, "No locations to search inside --search=%s\n", *searchFilename)
		os.Exit(1)
	}

	t0 := time.Now()
	if err := tracker.Load(*dataFilename); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load --input=%s: %+v\n", *dataFilename, err)
		os.Exit(1)
	}
	t1 := time.Now()
	log.Infof("Loaded tracks for %d unique vehicles ...", tracker.NrVehicles())
	log.Infof("Load data into memory took: %+v", t1.Sub(t0))
	t0 = t1

	//add last positions of all vehicles to the current grid
	//note: during load they are not added, because their may be many tracks for
	//the same vehicle, so if added during load will have to repeatedly add/remove
	//tracks for one vehicle, so here we only put the last track in the grid

	//make blocks of 1000 spread over lat and 1000 over lon ranges
	g, err := tracker.NewAtTs(time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create a grid: %+v", err)
		os.Exit(1)
	}

	//find closest vehicle at specified list of locations:
	t1 = time.Now()
	log.Infof("Grid construction at time now took: %+v", t1.Sub(t0))
	t0 = t1
	withGrid(searches, g)
	t1 = time.Now()
	log.Infof("Search took: %+v", t1.Sub(t0))
	jsonOutput, _ := json.Marshal(searches)
	fmt.Fprintln(os.Stdout, string(jsonOutput))
}

type Search struct {
	Loc     tracker.Loc      `json:"location"`
	Closest *tracker.Vehicle `json:"closest,omitempty"`
	//closestValue float64
}

// func simpleLoop(searches []Search, vehicles []*tracker.Vehicle) {
// 	tNow := time.Now()
// 	for _, v := range vehicles {
// 		for i, p := range searches {
// 			loc, err := v.LocAtTs(tNow)
// 			if err != nil {
// 				panic(err)
// 			}
// 			d := p.Loc.Diff(loc)
// 			distValue := math.Sqrt(float64(d.Lat)*float64(d.Lat) + float64(d.Lon)*float64(d.Lon))
// 			if p.Closest == nil || distValue < p.closestValue {
// 				searches[i].Closest = v
// 				searches[i].closestValue = distValue
// 			}
// 		}
// 	}
// }

func withGrid(searches []Search, g *tracker.Grid) error {
	//make blocks of 1000 spread over lat and 1000 over lon ranges
	// g, err := grid.New()
	// if err != nil {
	// 	return errors.Wrapf(err, "failed to create a grid")
	// }
	//put each vehicle in the grid
	// for _, v := range vehicles {
	// 	g.Add(v)
	// }

	// log.Debugf("Added %d -> grid.NrVehicles=%d", len(vehicles), g.NrVehicles())

	//for each search location, search only the block that the vehicle is in
	//and the blocks arround it
	for i, s := range searches {
		searches[i].Closest = g.FindClosest(s.Loc)
	}
	return nil
}
