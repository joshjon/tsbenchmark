#!/bin/bash

while [ "$(docker logs tsbenchmark_timescaledb_1 2>&1 | grep -c ' database system is ready to accept connections')" -eq "0" ]; do
  sleep 1
done
