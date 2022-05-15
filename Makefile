.PHONY: up down build run unit smoke

up:
	docker-compose -p tsbenchmark up -d
	./wait-for-ts.sh

down:
	docker-compose -p tsbenchmark down


build:
	docker build -t local/tsbenchmark .

run:
	docker run --rm --name tsbenchmark \
 		--network tsbenchmark_default \
 		--volume ${CURDIR}/database/query_params.csv:/data/query_params.csv \
 		local/tsbenchmark  -m 5 /data/query_params.csv

unit:
	go test -count=1 ./...

smoke:
	go test --tags=smoke -count=1 ./cmd...
