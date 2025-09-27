.PHONY: run build test clean

run: 
	go run main.go

build: 
	go build -o bin/app main.go

test: 
	go test  -bench=. ./... -v

cover: 
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -rf bin/app

sast:
	gosec ./...
# 	gitleaks git -v

dast: 
	docker run --rm ghcr.io/sullo/nikto -h http://host.docker.internal:8080

lint:
	go vet ./...
	go fmt ./...


curl: 
	curl http://localhost:8080/sum?a=5\&b=10
	echo "\n"
	curl http://localhost:8080/sum?a=5
	echo "\n"
	curl http://localhost:8080/sum?a=9223372036854775807\&b=100