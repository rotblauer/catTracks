#!/usr/bin/env bash

set -x

PORT=${PORT:-3001}

TDATA_CANON=${HOME}/tdata

TDATA_TEMP=${TDATA_TEMP:-$(mktemp -d)}

rm -rf ${TDATA_TEMP}
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
