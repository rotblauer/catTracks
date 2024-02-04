#!/usr/bin/env bash

set -e
set -x

for name in "rye" "ia"; do
  cat ~/tdata/lastpushes/$name.json | jj -p > testdata/$name.json

  zcat ~/tdata/edge.json.gz | catnames-cli modify | grep $name | tail -100  > testdata/$name.ndgeojson
  zcat ~/tdata/edge.json.gz | catnames-cli modify | grep $name | tail -100 | ndgeojson2geojsonfc > testdata/$name.geojsonfc
done

cp -a testdata/* ../catTrackslib/testdata/
