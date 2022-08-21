package tracker

import (
	"fmt"
	"math"
	"time"
)

//place all vehicles at specified timestamp in a grid
//if ts == 0, use latest location
func NewAtTs(ts time.Time) (*Grid, error) {
	g := &Grid{
		topBlock: nil, //everything is in this block
	}

	for _, v := range vehiclesByReg {
		loc, err := v.LocAtTs(ts)
		if err != nil {
			log.Errorf("%v", err)
			continue //skip this vehicle
		}
		if g.topBlock == nil {
			//first vehicle, create the first block
			b := &Block{
				size:          1, //default 1 degree
				minLat:        math.Floor(loc.Lat),
				minLon:        math.Floor(loc.Lon),
				children:      nil,
				parent:        nil,
				vehicleAtLocs: []vehicleAtLoc{{v, loc}},
			}
			b.midLat = b.minLat + 0.5*b.size
			b.midLon = b.minLon + 0.5*b.size
			b.maxLat = b.minLat + b.size
			b.maxLon = b.minLon + b.size
			g.topBlock = b
			//log.Debugf("Created first block: %+v", *b)
		} else {
			//add to existing topBlock, potentially replacing top with new layer
			g.topBlock = g.topBlock.Add(v, loc)
		}
		//log.Debugf("Add(%+v)->top=%+v", v, g.topBlock)
	} //for each vehicle
	return g, nil
} //NewAtTs()

type Grid struct {
	topBlock *Block
}

func (g Grid) FindClosest(searchLoc Loc) *Vehicle {
	if g.topBlock == nil {
		return nil
	}

	log.Infof("Grid: lat:%f..%f, lon:%f..%f",
		g.topBlock.minLat, g.topBlock.maxLat,
		g.topBlock.minLon, g.topBlock.maxLon,
	)

	log.Debugf("=====[ FIND %+v ]=====", searchLoc)
	//find block where search location is
	b := g.topBlock.ChildByLoc(searchLoc)
	log.Infof("Start in %+v with %d vehicles ...", b, b.NrVehicles())
	v := b.FindClosest(searchLoc)
	log.Infof("  FOUND %+v close to %+v", v, searchLoc)
	return v
}

func (g Grid) NrGridVehicles() int {
	if g.topBlock == nil {
		return 0
	}
	return g.topBlock.NrVehicles()
}

type Block struct {
	size                   float64
	minLat, midLat, maxLat float64
	minLon, midLon, maxLon float64
	children               [][]*Block //[0..1 for lat][0..1 for lon]
	parent                 *Block
	vehicleAtLocs          []vehicleAtLoc
}

type vehicleAtLoc struct {
	vehicle *Vehicle
	loc     Loc
}

//returns top block where added
func (b *Block) Add(v *Vehicle, loc Loc) *Block {
	if loc.Lat < b.minLat ||
		loc.Lat >= b.maxLat ||
		loc.Lon < b.minLon ||
		loc.Lon >= b.maxLon {
		return b.Parent(loc).Add(v, loc)
	}

	if b.children == nil {
		//no smaller cells, add to self
		if b.vehicleAtLocs == nil {
			b.vehicleAtLocs = []vehicleAtLoc{{v, loc}}
		} else {
			b.vehicleAtLocs = append(b.vehicleAtLocs, vehicleAtLoc{v, loc})
		}

		//if block gets busy, split into children
		if len(b.vehicleAtLocs) > 1000 { //limit nr of vehicles per block
			for _, val := range b.vehicleAtLocs {
				b.ChildByLoc(val.loc).Add(val.vehicle, val.loc)
			}
			b.vehicleAtLocs = nil
		}

		//log.Debugf("%+v.Added(%+v) ... #v=%d", *b, *v, len(b.vehicleAtLocs))
		return b.Top()
	}

	//has smaller cells, add to one of those
	return b.ChildByLoc(loc).Add(v, loc)
}

//Parent() makes a parent with twice own size if none exists and add this as a child
func (b *Block) Parent(loc Loc) *Block {
	if b.parent == nil {
		latIndex := 0
		lonIndex := 0
		p := &Block{
			size:          b.size * 2,
			minLat:        b.minLat,
			minLon:        b.minLon,
			parent:        nil,
			children:      [][]*Block{{nil, nil}, {nil, nil}},
			vehicleAtLocs: nil, //this is a parent, so vehicles will always be in children
		}
		if loc.Lat < b.minLat {
			latIndex = 1
			p.minLat -= b.size
		}
		if loc.Lon < b.minLon {
			lonIndex = 1
			p.minLon -= b.size
		}
		p.midLat = p.minLat + 0.5*p.size
		p.midLon = p.minLon + 0.5*p.size
		p.maxLat = p.minLat + p.size
		p.maxLon = p.minLon + p.size
		p.children[latIndex][lonIndex] = b
		b.parent = p
		//log.Debugf("  %+v<--[%d][%d]. (+)  %+v", b, latIndex, lonIndex, b.parent)
	}
	//log.Debugf("  %+v parent  <--  %+v", b.parent, b)
	return b.parent
}

func (b *Block) Top() *Block {
	if b.parent != nil {
		return b.parent.Top()
	}
	return b
}

//create child if not exist
func (b *Block) ChildByLoc(loc Loc) *Block {
	latIndex := 0
	//log.Debugf("%+v.mid(%8.3f,%8.3f).Child(%+v)...", b, b.midLat, b.midLon, loc)
	if loc.Lat >= b.midLat {
		latIndex = 1
	}
	lonIndex := 0
	if loc.Lon >= b.midLon {
		lonIndex = 1
	}

	if b.children == nil {
		b.children = [][]*Block{{nil, nil}, {nil, nil}}
	}
	if b.children[latIndex][lonIndex] == nil {
		sz := b.size * 0.5
		c := &Block{
			size:          sz,
			minLat:        b.minLat + sz*float64(latIndex),
			minLon:        b.minLon + sz*float64(lonIndex),
			parent:        b,
			children:      nil,
			vehicleAtLocs: nil,
		}
		c.midLat = c.minLat + 0.5*sz
		c.midLon = c.minLon + 0.5*sz
		c.maxLat = c.minLat + sz
		c.maxLon = c.minLon + sz
		b.children[latIndex][lonIndex] = c

		log.Debugf("  %+v.[%d][%d]--> (+)  %+v", b, latIndex, lonIndex, c)
	}
	//log.Debugf("  %+v.[%d][%d]  -->  %+v", b, latIndex, lonIndex, b.children[latIndex][lonIndex])
	c := b.children[latIndex][lonIndex]
	if c.children != nil {
		return c.ChildByLoc(loc)
	}
	return c
}

func (b Block) Format(f fmt.State, c rune) {
	//	f.Write([]byte(fmt.Sprintf("B(%8.3f..%8.3f..%8.3f;%8.3f..%8.3f..%8.3f)", b.minLat, b.midLat, b.maxLat, b.minLon, b.midLon, b.maxLon)))
	f.Write([]byte(fmt.Sprintf("B(%8.3f..%8.3f;%8.3f..%8.3f)", b.minLat, b.maxLat, b.minLon, b.maxLon)))
}

func (b Block) NrVehicles() int {
	nr := len(b.vehicleAtLocs)
	if b.children != nil {
		for latIndex := 0; latIndex < 2; latIndex++ {
			for lonIndex := 0; lonIndex < 2; lonIndex++ {
				c := b.children[latIndex][lonIndex]
				if c != nil {
					nr += c.NrVehicles()
				}
			}
		}
	}
	return nr
}

func (b *Block) FindClosest(loc Loc) *Vehicle {
	if b.NrVehicles() == 0 {
		if b.parent == nil {
			return nil
		}
		return b.parent.FindClosest(loc)
	}

	//has some vehicles in this block or its childrent
	//see which are the closest
	//NOTE: Not 100% fool proof around edge cases of the block
	//	as may have one vehicle in far corner and neightbour has one near our edge,
	//  then the latter will not be reckoned

	//check own vechicles
	cv, _ := b.closest(loc, nil, 0)
	return cv
}

//look for closest in self and children - not considering parent
func (b Block) closest(loc Loc, cv *Vehicle, cd float64) (*Vehicle, float64) {
	for _, val := range b.vehicleAtLocs {
		dist := loc.Dist(val.loc)
		log.Debugf("  %+v.dist = %f to %+v", b, dist, val.loc)
		if cv == nil || dist < cd {
			cv = val.vehicle
			cd = dist
		}
	}

	//check all child vehicles
	if b.children != nil {
		for latIndex := 0; latIndex < 2; latIndex++ {
			for lonIndex := 0; lonIndex < 2; lonIndex++ {
				c := b.children[latIndex][lonIndex]
				if c != nil {
					cv, cd = c.closest(loc, cv, cd)
				}
			}
		}
	}
	return cv, cd
}
