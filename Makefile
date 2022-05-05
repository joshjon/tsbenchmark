.PHONY: db-up run db-down build run unit

db-up:
	docker-compose -p timescaledb up -d --build

db-down:
	docker-compose -p timescaledb down


build:
	docker build -t local/tsbenchmark .

run:
	docker run --rm --name tsbenchmark local/tsbenchmark foo.csv -m 10 -s 1000 -w 10000

unit:
	go test -count=1 ./...
