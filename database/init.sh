#!/bin/bash
echo "loading cpu usage data"
psql -U postgres -d homework -c "\COPY cpu_usage FROM /var/lib/postgresql/data/cpu_usage.csv CSV HEADER"
