#!/usr/bin/env bash

TDATA_ROOT=${HOME}/tdata-tests

env COTOKEN=mytoken \
  catTracks \
    --port 3001 \
    --db-path-master ${TDATA_ROOT}/tracks.db \
    --db-path-devop ${TDATA_ROOT}/devop.db \
    --db-path-edge ${TDATA_ROOT}/edge.db \
    --tracks-gz-path ${TDATA_ROOT}/master.json.gz \
    --devop-gz-path ${TDATA_ROOT}/devop.json.gz \
    --edge-gz-path ${TDATA_ROOT}/edge.json.gz \
    --master-lock ${TDATA_ROOT}/MASTERLOCK \
    --devop-lock ${TDATA_ROOT}/DEVOPLOCK \
    --edge-lock ${TDATA_ROOT}/EDGELOCK \
    --proc-master \
    --proc-edge
