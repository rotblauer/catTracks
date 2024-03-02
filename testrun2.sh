#!/usr/bin/env bash

set -e
set -x

trap 'pkill -P $$; exit' SIGINT SIGTERM 

{ env PORT=3001 ./testrun.sh --forward-url 'http://localhost:3002/populate/' |& tee /tmp/cattracks1.log ; } &
{ env PORT=3002 ./testrun.sh |& tee /tmp/cattracks2.log ; } &

wait

