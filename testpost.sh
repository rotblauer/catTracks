#!/usr/bin/env bash

set -e
set -x

curl -X POST \
    -H "Content-Type: application/json" \
    -H "AuthorizationOfCats: mytoken" \
    -d @testdata/trackpoints.json http://localhost:3001/populate/

curl -X POST \
    -H "Content-Type: application/json" \
    -H "AuthorizationOfCats: mytoken" \
    -d @testdata/lastknown-rye.geojson http://localhost:3001/populate/

curl -X POST \
    -H "Content-Type: application/json" \
    -H "AuthorizationOfCats: mytoken" \
    -d @testdata/lastknown-ia.geojson http://localhost:3001/populate/
