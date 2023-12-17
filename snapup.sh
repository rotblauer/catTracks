#!/usr/bin/env bash

set -x

BUCKET=rotblauercatsnaps
TARGET=${1:-/home/ia/tsnaps/aws/${BUCKET}}

mkdir -p ${TARGET}

# https://stackoverflow.com/questions/8659382/downloading-an-entire-s3-bucket
aws s3 sync s3://${BUCKET} ${TARGET}
