#!/bin/bash

while [ "$(docker logs tsbenchmark-timescaledb-1 2>&1 | grep -c 'ready to accept connections')" -eq "0" ]; do
  sleep 1
done
