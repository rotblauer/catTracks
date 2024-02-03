#!/usr/bin/env bash

set -e
set -x

PORT=${PORT:-3001}

curl -X POST \
    -H "Content-Type: application/json" \
    -H "AuthorizationOfCats: mytoken" \
    -d @testdata/ia.geojsonfc \
    http://localhost:3001/populate/

curl -X POST \
    -H "Content-Type: application/json" \
    -H "AuthorizationOfCats: mytoken" \
    -d @testdata/rye.geojsonfc \
    http://localhost:3001/populate


# ! FIXME: This doesn't work:
#
# 2024/02/02 09:04:59 POST /populate/ HTTP/1.1
# Host: localhost:3001
# Accept: */*
# Authorizationofcats: mytoken
# Content-Length: 70992
# Content-Type: application/json
# Expect: 100-continue
# User-Agent: curl/7.68.0
#
#
# 2024/02/02 09:04:59 Decoding 70992 bytes
# 2024/02/02 09:04:59 attempting decode as ndjson instead..., length: 2 []
# 2024/02/02 09:04:59 Decoded 0 features
# 2024/02/02 09:04:59 POST /populate/ HTTP/1.1
# Host: localhost:3001
# Accept: */*
# Authorizationofcats: mytoken
# Content-Length: 65506
# Content-Type: application/json
# Expect: 100-continue
# User-Agent: curl/7.68.0
#
#
# 2024/02/02 09:04:59 Decoding 65506 bytes
# 2024/02/02 09:04:59 attempting decode as ndjson instead..., length: 2 []
# 2024/02/02 09:04:59 Decoded 0 features
#
# curl -X POST \
#     -H "Content-Type: application/json" \
#     -H "AuthorizationOfCats: mytoken" \
#     -d @testdata/ia.ndgeojson http://localhost:3001/populate/
#
# curl -X POST \
#     -H "Content-Type: application/json" \
#     -H "AuthorizationOfCats: mytoken" \
#     -d @testdata/rye.ndgeojson http://localhost:3001/populate/



# curl -X POST \
#     -H "Content-Type: application/json" \
#     -H "AuthorizationOfCats: mytoken" \
#     -d @testdata/trackpoints.json http://localhost:3001/populate/
#
# curl -X POST \
#     -H "Content-Type: application/json" \
#     -H "AuthorizationOfCats: mytoken" \
#     -d @testdata/lastknown-rye.geojson http://localhost:3001/populate/
#
# curl -X POST \
#     -H "Content-Type: application/json" \
#     -H "AuthorizationOfCats: mytoken" \
#     -d @testdata/lastknown-ia.geojson http://localhost:3001/populate/
