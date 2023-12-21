
cattracks 2023
a year in report


december 2023
cattracks gets updated. the v1.1+ series.
cattracks was a monster.
tippecanoe _is_ a monster.
i don't know how long the procmaster runs were, but i'm sure they were over 24 hours.
tracks.db was 335GB and counting.
225 million tracks, starting properly sometime in... 2018? (earlier tracks of 2010 derived from photos geostamps or other sources like bike gps tracks)
but the master.json.gz was only 5.6GB. 
mbtiles around the <10GB mark.
johnny sparked it all as usual by creating geocat, a genomics pipeline for cattracks
i wrote a program to tracks cats by cat square meters (cattracks-split-uniqcell-gz)
it uses an lru cache in front of a bbolt db index on a S2's cell id, at a level corresponding to approximately 1 square meter (level 23).
the new cattracks uses this for a couple things.
it generates .mbtiles files for each cat (eg. ia.level-23.mbtiles),
and only regenerates them when the cats posts fresh powder tracks.
rye.json.gz -> rye.level-23.mbtiles takes about 20 minutes on the ol' thinkpad, about 63 minutes on rottor.
rye.json.gz 500MB, ia.json.gz about 300MB; mbtiles both under 3GB.
there's a really horrible decision i made that uses the despicable idea of the genpop vs the cats that rule the world,
so the genpop basically gets kind of fucked unil i fix catonmap.net.
so unless the cat pushes powder, tippe sleeps. or tippe runs cat lifetime unique square meter tracks.
not ideal still, obviously.
there's a compressed tracks.db.bak_20231218.gz on rottor. get it while its still in style. pigz ftw.


what cattracks ralllly needs to do is to generate tiles only when needed,
and to generate only the tiles that are needed.
it needs to become the tile server.
tippe already makes great annotations, like 1000 attribute values stored, all fields, theres a schema.
so you have context. and we have indexable cell ids, so maybe an area-contains query wouldn't take thaaaaat long, and you could cache everything (~5GB maybe)
tippecanoe's tile-join, what does 'merge tiles' mean? 

- save the snaps (S3 reliant now, install --snaps-output=...)
- save the tracks
 - `cattracks --tracks-output=master_strict.json.gz,edge.json.gz`
- serve the maps
 - catona: put cats on tiles
- serve the stats

---

- [ ] cattracks-names: package provides cat aliases for cat names
- [ ] cattracks-structs: package provides data structures for cat tracks
- [ ] cattracks-m2: 
