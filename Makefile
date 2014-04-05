packages = common network quorum disk main common/erasure common/crypto
testpackages = $(addsuffix /..., $(packages))

all: submodule-update libraries

submodule-update:
	git submodule update

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

.PHONY: all test fmt libraries test-verbose docs submodule-update
