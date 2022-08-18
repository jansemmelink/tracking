# Tracking

The source file is not part of this archive.
Copy it into ./data/:
```
$ ls -l data
-rw-r--r--  1 jans  staff  60000000 Aug 18 22:54 VehiclePositions.dat
```

You can also load from other location with `--input=...` option.

# File Format
The spec you gave for the file format was a bit sparse...

Until just a few minute ago, I assumed "null terminated string" means I must read until I got to the end, i.e. could be any size. Only after some analysis on the data that I parsed, I realised ALL the registrations are 9 char long. Nice. At that point I changed the C code to just read into char[10] and it simplified the code and sped it up too. This also meant that I could work with a fixed record structure which I then did in Go, reading 30 bytes at a time. I could probably change it to Read like 1000x30 which will make it faster, or do 4096 byte buffers as in C and shift the remaining data before reading the next buffer. I think that is still the fastest way to read it.

Also after completing the task and analysing the file I realised that the same vehicles appears many times in the file. I worked from the assumption that you were tracking 2M vehicles and the file was a snapshot of the last time they reported their location... obviously not. So in my code, I am not yet removing duplicates. I just added all 2M vehicles to the grid. I do not recall you asking for the closest vehicle at a specific time...


# C Code
In the C directory is just code to load the file into memory.
It does 4096 byte block reads, process complete records then shift and read more data. It is pretty quick to read the whole file, 0.11s on my MacBook with M1 CPU and SSD:
```
% cd c
% gcc main.c -o tracking
% time ./tracking
loaded 2000000 entries
./tracking  0.11s user 0.02s system 97% cpu 0.133 total
```
I built it on Centos6 and MacOS and it should work on Ubuntu as well or WSL2 for Windows, using the gcc compiler.

# Golang Code
In the main directory is a GoLang program to also ready the file and then do the searches:
```
% go get .
% go build
% time ./tracking
...
2022-08-19 00:29:15.516 DEBUG            main.go(   58): Load and grid took: 4.988404584s map[]
2022-08-19 00:29:15.517 DEBUG            main.go(   61): Search took: 788.666µs map[]
...
./tracking  13.19s user 0.82s system 278% cpu 5.030 total
```

It loads significantly slower than C, because:
- It reads 30 bytes at a time from the file, then parse it before it reads again.
- It also adds to the grid as they are read, constructing more blocks as needed.

So here the load and construction of the grid took 4 seconds
But after that was done, all 10 searches were completed in 788us

# Grid (in golang)

The grid consists of blocks.

When the first vehicle is added, the first block is automatically created to contain that vehicle. The first block has size of 1 degree in lat and lon directions.

When the next vehicle is created, it is added to the same block if it fits in the boundaries, else it is added to the parent using recursion. The parent is automatically constructed to cover twice the area of the child where the vehicle did not fit, and it will expand in the direction of the new vehicle. Recursion will continue making bigger parents until the new vehicle fit.

Adding parent replace the top element of the grid, to next vehicles are added from that point.

Apart from covering an area between minLat..maxLat and minLon..maxLon,
each block has a list of vehicles, and new vehicles are added to the list if they fall in this boundary.

Once the block reaches a threshold (set at 1000 for now), the block automatically break into 4 quadrants moving all of its own vehicles into one of those, and from this point on, vehicles added to this block will be added to one of the quadrants.

The quadrants are also blocks, so will also break into smaller blocks when they reach the threshold.

So blocks expand automatically up and down as vehicles are added.

After all vehicles were added, only can run grid.NrVehicles which counts then with recursion to confirm all 2,000,000 vehicles from the file were added and none were duplicated.

# Searching

Searches start from the top block that spans the whole area in:
```
g.FindClosest(searchLocation)
```

The search starts by finding the block where the search location resides.
Then it search only in that block.

For example:
```
grid.go(   62): Start in B(  34.500..  34.625;-102.125..-102.000) with 991 vehicles ...
grid.go(   64):   FOUND V(8Y-044 SP:(  34.545;-102.102)) close to (  34.545;-102.101)
```

I do not believe that is completely 100% the closest vehicle, because I suspect there are a few edge cases where say a block has one vehicle in the one corner and we search in the opposite corner of the same block. Then that vehicle will appear the closest. Now if another vehicle sits just outside my block in the neighbouring block, then it is actually closer but not yet found with this algorithm.

To solve that, I should search neightbouring blocks to the left/right/top/bottom and diagonally from the search location too... With blocks having different granularity, the .Left() and .Right() and .Up() and .Down() ...and DiagonalXXX() functions will also have to recurse to the parent's children.

I am just running out of time, so am going to stop here.

It was a fun exercise so far, but I cannot afford to spend more time on this right now.

I cleaned the code somewhat, but it is obviously not super clean yet.

Regards,
Jan
