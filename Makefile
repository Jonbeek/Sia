packages = common network quorum disk main common/erasure
testpackages = $(addsuffix /..., $(packages))

all: libraries

race-libs:
	GOPATH=$(CURDIR) go install -race std

libraries: src/*/*.go fmt
	GOPATH=$(CURDIR) go install $(packages)

test: libraries
	GOPATH=$(CURDIR) go test $(testpackages)

bench: libraries
	GOPATH=$(CURDIR) go test $(packages)

test-verbose: libraries
	GOPATH=$(CURDIR) go test -test.v $(testpackages)

fmt:
	GOPATH=$(CURDIR) go fmt $(packages)

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all test fmt libraries test-verbose docs
