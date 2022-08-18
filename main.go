package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/jansemmelink/tracking/grid"
	"github.com/jansemmelink/tracking/tracker"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func main() {
	limit := flag.Int("limit", 0, "Limit nr of vehicles to process (default 0 = all in file)")
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

	//make blocks of 1000 spread over lat and 1000 over lon ranges
	g, err := grid.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create a grid: %+v", err)
		os.Exit(1)
	}

	t0 := time.Now()
	if err := tracker.Load(*dataFilename, *limit, g); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load --input=%s: %+v\n", *dataFilename, err)
		os.Exit(1)
	}
	//start searching:
	t1 := time.Now()
	log.Debugf("Load and grid took: %+v", t1.Sub(t0))
	withGrid(searches, g)
	t2 := time.Now()
	log.Debugf("Search took: %+v", t2.Sub(t1))

	jsonOutput, _ := json.Marshal(searches)
	fmt.Fprintln(os.Stdout, string(jsonOutput))
}

type Search struct {
	Loc          tracker.Loc      `json:"location"`
	Closest      *tracker.Vehicle `json:"closest,omitempty"`
	closestValue float64
}

func simpleLoop(searches []Search, vehicles []*tracker.Vehicle) {
	for _, v := range vehicles {
		for i, p := range searches {
			d := p.Loc.Diff(v.Loc)
			distValue := math.Sqrt(float64(d.Lat)*float64(d.Lat) + float64(d.Lon)*float64(d.Lon))
			if p.Closest == nil || distValue < p.closestValue {
				searches[i].Closest = v
				searches[i].closestValue = distValue
			}
		}
	}
}

func withGrid(searches []Search, g *grid.Grid) error {
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
