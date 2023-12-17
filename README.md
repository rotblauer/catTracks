

# catTracks

Tracking Cats and generating data

20231214

cattracks is too big, too heavy, and thus too expensive
while `master.json.gz` is only 6GB
master tippecanoe proc is the CPU+memory hog, b/c its tiling the whole world for all 250 million tracks every time 
tracks.db is the footprint hog, 355GB, storing all tracks uncompressed, and for no reason (we don't read or use the indexes, except for deduping points which almost-to-always is unnecessary)

things i want to change to make it lighter:

- tracks.db
  - important things:
    - "catsnaps" bucket stores 3k+ tracks with snaps
      - these tracks have .notes.imgS3
      - the actual images are stores in S3/rotblauercatsnaps
      - /catsnaps handler serves the catsnaps
    - /lastKnown handler uses
      - bucket="stats",key=lastknown returns all cats last known location, indexed on catname
      - `type LastKnown map[string]*trackPoint.TrackPoint`
- mbtiles generation
  - currently the procmaster tippe takes... a long time; 24hrs+, maybe even 48..72..96hrs+ (this is `master.json.gz -> master.mbtiles`)
  - i can run (w/ same tippe config) tippe on cat:uniqcells for all cats in 44m minutes on my laptop 

---
pre-202312

//TODO

- LAT/LONG :heavy_check_mark: db
- ELEVATION :heavy_check_mark: db
- SPEED :heavy_check_mark: db
- compass heading :heavy_check_mark: db
- TILT
- NUmber of cats
- cat heartrate :heavy_check_mark: db
- plots, mapper :heavy_check_mark:
- pics, first person and third person, upload if not black and with wifi and lots of battery

- First generator on computer desktop, then IOS.

//TODO , need to grap GAE datastore, or maybe just forgetaboutit

### plays well with [https://github.com/rotblauer/bildungs-roamin](https://github.com/rotblauer/bildungs-roamin)

### Starting development server

:beer:
```
go run main.go
```
### Frey bay bay

![does this](./example.png)
