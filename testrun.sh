#!/usr/bin/env bash

set -x

TDATA_ROOT=${HOME}/tdata
TDATA_ROOT_TEST=/tmp/tdata-tests

rm -rf ${TDATA_ROOT_TEST}
mkdir -p ${TDATA_ROOT_TEST}

# Copy the origin tracks.db because it contains cat snaps, which I want
# to check, but not post.
cp "${TDATA_ROOT}/tracks.db" "${TDATA_ROOT_TEST}/tracks.db"

env COTOKEN=mytoken \
./build/bin/cattracks \
    --port 3001 \
    --db-path-master ${TDATA_ROOT_TEST}/tracks.db \
    --db-path-devop ${TDATA_ROOT_TEST}/devop.db \
    --db-path-edge ${TDATA_ROOT_TEST}/edge.db \
    --tracks-gz-path ${TDATA_ROOT_TEST}/master.json.gz \
    --devop-gz-path ${TDATA_ROOT_TEST}/devop.json.gz \
    --edge-gz-path ${TDATA_ROOT_TEST}/edge.json.gz \
    --master-lock ${TDATA_ROOT_TEST}/MASTERLOCK \
    --devop-lock ${TDATA_ROOT_TEST}/DEVOPLOCK \
    --edge-lock ${TDATA_ROOT_TEST}/EDGELOCK \
    --proc-master \
    --proc-edge
