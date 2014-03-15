all: libraries

race-libs:
	GOPATH=$(CURDIR) go install -race std

libraries: src/*/*.go fmt
	GOPATH=$(CURDIR) go install network swarm disk common main

test: libraries
	GOPATH=$(CURDIR) go test network swarm disk common main -race -timeout 30s

bench: libraries
	GOPATH=$(CURDIR) go test network swarm disk common main -bench .

test-verbose: libraries
	GOPATH=$(CURDIR) go test -test.v network swarm disk common main

fmt:
	go fmt src/network/*.go
	go fmt src/swarm/*.go
	go fmt src/disk/*.go
	go fmt src/common/*.go
	go fmt src/main/*.go

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all test fmt libraries test-verbose docs
