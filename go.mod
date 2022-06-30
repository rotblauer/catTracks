module github.com/rotblauer/catTracks

go 1.12

require (
	github.com/kpawlik/geojson v0.0.0-20171201195549-1a4f120c6b41
	github.com/rotblauer/catTrackslib v1.0.1
	github.com/rotblauer/trackpoints/trackPoint v0.0.0-20220630172156-84e70c5d820e // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
)

replace (
	github.com/rotblauer/catTrackslib v1.0.1 => /home/ia/go/src/github.com/rotblauer/catTrackslib
)
