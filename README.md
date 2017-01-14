

# catTracks

Tracking Cats and generating data

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


### Starting development server

:beer:
```
brew install go-app-engine-64
goapp serve app.yml
# or reset dev datastore on swerver start
goapp serve -clear_datastore app.yaml
goapp serve .
```

### Uploading to google app engine

:beers:
```
brew install go-app-engine-64
goapp deploy
```
