#!/usr/bin/env bash

# This script provides a base configuration and setup for running cattracks
# on a development machine for testing.
# This script uses an ephemeral datadir seeded with data from the canonical data dir.
# Additional hardcoded configuration decisions:
# - 'mytoken' is used as the Cat Owners Token
# 
# Customized the run:
# - additional cattracks CLI flags can be passed as arguments ($*)
#
# USAGE:
#
#   env PORT=3001 ./testrun.sh --forward-url='http://localhost:3002' --disable-websocket
#   env PORT=3002 ./testrun.sh --disable-websocket

set -e
set -x

PORT=${PORT:-3001}

# TDATA_CANON is the canonical location of the actual cat tracks.
# This script copies the tracks.db to the ephemeral TDATA_TEMP
# because I want to test cattracks on top of real data.
TDATA_CANON=${HOME}/tdata

# TDATA_TEMP is the ephemeral location of the cat track datadir for this test instance.
TDATA_TEMP=${TDATA_TEMP:-$(mktemp -d)}
trap "rm -rf ${TDATA_TEMP}" EXIT
mkdir -p ${TDATA_TEMP}

# Copy the origin tracks.db because it contains cat snaps, which I want
# to check, but not post.
cp "${TDATA_CANON}/tracks.db" "${TDATA_TEMP}/tracks.db"

mkdir -p ./build/bin
go build -o ./build/bin/cattracks .

env COTOKEN=mytoken \
./build/bin/cattracks \
    --port ${PORT} \
    --db-path-master ${TDATA_TEMP}/tracks.db \
    --db-path-devop ${TDATA_TEMP}/devop.db \
    --db-path-edge ${TDATA_TEMP}/edge.db \
    --tracks-gz-path ${TDATA_TEMP}/master.json.gz \
    --devop-gz-path ${TDATA_TEMP}/devop.json.gz \
    --edge-gz-path ${TDATA_TEMP}/edge.json.gz \
    --master-lock ${TDATA_TEMP}/MASTERLOCK \
    --devop-lock ${TDATA_TEMP}/DEVOPLOCK \
    --edge-lock ${TDATA_TEMP}/EDGELOCK \
    --proc-master \
    --proc-edge $*
