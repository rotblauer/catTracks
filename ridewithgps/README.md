When Global Cat Positioning System fails, but there's a RideWithGPS alternative.
I want to import my ride tracks, recorded by my Wahoo Elemnt Roam, into Cat Tracks. 

RideWithGPS exported my ride to a GPX file `07_17_24.gpx`.

I converted the GPX data to GeoJSON using [togeojson](https://mapbox.github.io/togeojson/) `07_17_24.geojson`.

https://mapbox.github.io/togeojson/
https://github.com/placemark/togeojson

This GeoJSON file must now be coerced into the Cat Tracks format.

```bash
cat ridewithgps/exports/07_17_24.geojson | go run ridewithgps/main.go > ridewithgps/07_17_24_cattracks.geojson
cat ridewithgps/exports/07_26_24.geojson | go run ridewithgps/main.go > ridewithgps/07_26_24_cattracks.geojson
```

This GeoJSON file `ridewithgps/07_17_24_cattracks.geojson` is now ready to be imported into Cat Tracks.

```bash
curl -X POST -H "Content-Type: application/json" -H "AuthorizationOfCats: thecattokenthatunlockstheworldtrackyourcats" -d @ridewithgps/07_17_24_cattracks.geojson 'http://track.areteh.co:3001/populate/'
curl -X POST -H "Content-Type: application/json" -H "AuthorizationOfCats: thecattokenthatunlockstheworldtrackyourcats" -d @ridewithgps/07_26_24_cattracks.geojson 'http://track.areteh.co:3001/populate/'
```

