#!/usr/bin/env bash

go run main.go --exportPostGIS --db-path-master ~/SAMSUNG_T5_3/tracksdata/tracks.db --export.target=postgres://postgres:mysecretpassword@localhost:5432/cattracks1?sslmode=disable

