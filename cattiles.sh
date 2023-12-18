#!/usr/bin/env bash

DATA_ROOT=~/tdata

MASTERGZ=${DATA_ROOT}/master.json.gz
EDGEGZ=${DATA_ROOT}/edge.json.gz
CATGZS_ROOT=${DATA_ROOT}/catgzs

# split master.json.gz into cat.json.gz if this has not been done yet before

# split edge.json.gz into cat.json.gz
# this saves time by not having to re-split master.json.gz

# then, append edge.json.gz to master.json.gz

# run cat tile maker on cat.json.gz

if [[ ! -d ${CATGZS_ROOT} ]]; then
    mkdir -p ${CATGZS_ROOT}

fi

zcat $EDGEGZ | go-cat-cells --workers 8 --batch-size 10000 --cell-level 23 --output ${CATGZS_ROOT}
