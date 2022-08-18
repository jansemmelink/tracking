package grid

import (
	"fmt"
	"math"

	"github.com/go-msvc/errors"
	"github.com/jansemmelink/tracking/tracker"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelInfo)

func New() (*Grid, error) {
	g := &Grid{
		topBlock: nil, //everything is in this block
	}
	return g, nil
} //New()

type Grid struct {
	topBlock *Block
}

func (g *Grid) Add(v *tracker.Vehicle) error {
	log.Debugf("ADD: %+v", v)
	if v.Loc.Lat < -90 || v.Loc.Lat > 90 || v.Loc.Lon < -180 || v.Loc.Lon > 180 {
		return errors.Errorf("invalid %+v", v)
	}
	if g.topBlock == nil {
		//first vehicle, create the first block
		b := &Block{
			size:     1, //default 1 degree
			minLat:   math.Floor(v.Loc.Lat),
			minLon:   math.Floor(v.Loc.Lon),
			children: nil,
			parent:   nil,
			vehicles: []*tracker.Vehicle{v},
		}
		b.midLat = b.minLat + 0.5*b.size
		b.midLon = b.minLon + 0.5*b.size
		b.maxLat = b.minLat + b.size
		b.maxLon = b.minLon + b.size
		g.topBlock = b
		log.Debugf("Created first block: %+v", *b)
	} else {
		//add to existing topBlock, potentially replacing top with new layer
		g.topBlock = g.topBlock.Add(v)
	}
	log.Debugf("Add(%+v)->top=%+v", v, g.topBlock)
	return nil
} //grid.Add()

func (g Grid) FindClosest(searchLoc tracker.Loc) *tracker.Vehicle {
	if g.topBlock == nil {
		return nil
	}

	log.Debugf("=====[ FIND %+v ]=====", searchLoc)
	//find block where search location is
	b := g.topBlock.ChildByLoc(searchLoc)
	log.Infof("Start in %+v with %d vehicles ...", b, b.NrVehicles())
	v := b.FindClosest(searchLoc)
	log.Infof("  FOUND %+v close to %+v", v, searchLoc)
	return v
}

func (g Grid) NrVehicles() int {
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
	vehicles               []*tracker.Vehicle
}

//returns top block where added
func (b *Block) Add(v *tracker.Vehicle) *Block {
	if v.Loc.Lat < b.minLat ||
		v.Loc.Lat >= b.maxLat ||
		v.Loc.Lon < b.minLon ||
		v.Loc.Lon >= b.maxLon {
		return b.Parent(v.Loc).Add(v)
	}

	if b.children == nil {
		if b.vehicles == nil {
			b.vehicles = []*tracker.Vehicle{v}
		} else {
			b.vehicles = append(b.vehicles, v)
		}

		//if block gets busy, split into children
		if len(b.vehicles) > 1000 { //limit nr of vehicles per block
			for _, v := range b.vehicles {
				b.ChildByLoc(v.Loc).Add(v)
			}
			b.vehicles = nil
		}

		log.Debugf("%+v.Added(%+v) ... #v=%d", *b, *v, len(b.vehicles))
		return b.Top()
	}
	return b.ChildByLoc(v.Loc).Add(v)
}

//Parent() makes a parent with twice own size if none exists and add this as a child
func (b *Block) Parent(l tracker.Loc) *Block {
	if b.parent == nil {
		latIndex := 0
		lonIndex := 0
		p := &Block{
			size:     b.size * 2,
			minLat:   b.minLat,
			minLon:   b.minLon,
			parent:   nil,
			children: [][]*Block{{nil, nil}, {nil, nil}},
			vehicles: nil, //this is a parent, so vehicles will always be in children
		}
		if l.Lat < b.minLat {
			latIndex = 1
			p.minLat -= b.size
		}
		if l.Lon < b.minLon {
			lonIndex = 1
			p.minLon -= b.size
		}
		p.midLat = p.minLat + 0.5*p.size
		p.midLon = p.minLon + 0.5*p.size
		p.maxLat = p.minLat + p.size
		p.maxLon = p.minLon + p.size
		p.children[latIndex][lonIndex] = b
		b.parent = p
		log.Debugf("  %+v<--[%d][%d]. (+)  %+v", b, latIndex, lonIndex, b.parent)
	}
	log.Debugf("  %+v parent  <--  %+v", b.parent, b)
	return b.parent
}

func (b *Block) Top() *Block {
	if b.parent != nil {
		return b.parent.Top()
	}
	return b
}

//create child if not exist
func (b *Block) ChildByLoc(l tracker.Loc) *Block {
	latIndex := 0
	log.Debugf("%+v.mid(%8.3f,%8.3f).Child(%+v)...", b, b.midLat, b.midLon, l)
	if l.Lat >= b.midLat {
		latIndex = 1
	}
	lonIndex := 0
	if l.Lon >= b.midLon {
		lonIndex = 1
	}

	if b.children == nil {
		b.children = [][]*Block{{nil, nil}, {nil, nil}}
	}
	if b.children[latIndex][lonIndex] == nil {
		sz := b.size * 0.5
		c := &Block{
			size:     sz,
			minLat:   b.minLat + sz*float64(latIndex),
			minLon:   b.minLon + sz*float64(lonIndex),
			parent:   b,
			children: nil,
			vehicles: nil,
		}
		c.midLat = c.minLat + 0.5*sz
		c.midLon = c.minLon + 0.5*sz
		c.maxLat = c.minLat + sz
		c.maxLon = c.minLon + sz
		b.children[latIndex][lonIndex] = c

		log.Debugf("  %+v.[%d][%d]--> (+)  %+v", b, latIndex, lonIndex, c)
	}
	log.Debugf("  %+v.[%d][%d]  -->  %+v", b, latIndex, lonIndex, b.children[latIndex][lonIndex])
	c := b.children[latIndex][lonIndex]
	if c.children != nil {
		return c.ChildByLoc(l)
	}
	return c
}

func (b Block) Format(f fmt.State, c rune) {
	//	f.Write([]byte(fmt.Sprintf("B(%8.3f..%8.3f..%8.3f;%8.3f..%8.3f..%8.3f)", b.minLat, b.midLat, b.maxLat, b.minLon, b.midLon, b.maxLon)))
	f.Write([]byte(fmt.Sprintf("B(%8.3f..%8.3f;%8.3f..%8.3f)", b.minLat, b.maxLat, b.minLon, b.maxLon)))
}

func (b Block) NrVehicles() int {
	nr := len(b.vehicles)
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

func (b *Block) FindClosest(l tracker.Loc) *tracker.Vehicle {
	if b.NrVehicles() == 0 {
		if b.parent == nil {
			return nil
		}
		return b.parent.FindClosest(l)
	}

	//has some vehicles in this block or its childrent
	//see which are the closest
	//NOTE: Not 100% fool proof around edge cases of the block
	//	as may have one vehicle in far corner and neightbour has one near our edge,
	//  then the latter will not be reckoned

	//check own vechicles
	cv, _ := b.closest(l, nil, 0)
	return cv
}

//look for closest in self and children - not considering parent
func (b Block) closest(l tracker.Loc, cv *tracker.Vehicle, cd float64) (*tracker.Vehicle, float64) {
	for _, v := range b.vehicles {
		dist := l.Dist(v.Loc)
		log.Debugf("  %+v.dist = %f to %+v", b, dist, v.Loc)
		if cv == nil || dist < cd {
			cv = v
			cd = dist
		}
	}

	//check all child vehicles
	if b.children != nil {
		for latIndex := 0; latIndex < 2; latIndex++ {
			for lonIndex := 0; lonIndex < 2; lonIndex++ {
				c := b.children[latIndex][lonIndex]
				if c != nil {
					cv, cd = c.closest(l, cv, cd)
				}
			}
		}
	}
	return cv, cd
}
