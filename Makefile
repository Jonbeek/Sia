all: libraries

libraries: src/*/*.go fmt
	GOPATH=$(CURDIR) go install network swarm disk common

test: libraries
	GOPATH=$(CURDIR) go test network swarm disk common

test-verbose:
	GOPATH=$(CURDIR) go test -test.v network swarm disk common

fmt:
	go fmt src/network/*.go
	go fmt src/swarm/*.go
	go fmt src/disk/*.go
	go fmt src/common/*.go

.PHONY: all test fmt libraries test-verbose
