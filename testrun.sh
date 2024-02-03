#!/usr/bin/env bash

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
