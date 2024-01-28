#!/usr/bin/env bash

curl -X POST \
    -H "Content-Type: application/json" \
    -H "AuthorizationOfCats: mytoken" \
    -d @testdata/trackpoints.json http://localhost:3001/populate/
