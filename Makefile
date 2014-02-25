all: libraries

race-libs:
	GOPATH=$(CURDIR) go install -race std

libraries: src/*/*.go fmt
	GOPATH=$(CURDIR) go install network swarm disk common main

test: libraries
	GOPATH=$(CURDIR) go test network swarm disk common main -race -timeout 5s

test-verbose: libraries
	GOPATH=$(CURDIR) go test -test.v network swarm disk common main

fmt:
	go fmt src/network/*.go
	go fmt src/swarm/*.go
	go fmt src/disk/*.go
	go fmt src/common/*.go
	go fmt src/main/*.go

.PHONY: all test fmt libraries test-verbose
