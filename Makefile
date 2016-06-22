.PHONY: build doc fmt lint run test vet
export GO15VENDOREXPERIMENT=1
default: build
build: fmt 
	go build -v -o omega-metrics ./
doc:
	godoc -http=:6060 -index
run: build
	./omega-metrics
fmt:
	go fmt ./...
test:
	go test ./...
