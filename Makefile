.PHONY: up down build run unit

up:
	docker-compose -p tsbenchmark up -d --build --force-recreate
	./wait-for-ts.sh

down:
	docker-compose -p tsbenchmark down


build:
	docker build -t local/tsbenchmark .

run:
	docker run --rm --name tsbenchmark local/tsbenchmark foo.csv -m 10 -s 1000 -w 10000

unit:
	go test -count=1 ./...
