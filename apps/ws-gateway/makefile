.PHONY: run build test clean

run:
	go run cmd/server/main.go

build:
	go build -o bin/ws-gateway cmd/server/main.go

test:
	go test ./...

clean:
	rm -rf bin/

docker-build:
	docker build -t nabo/ws-gateway .

docker-run:
	docker run -p 8080:8080 nabo/ws-gateway
