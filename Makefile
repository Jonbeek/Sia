gopath = GOPATH=$(CURDIR)
cgo_ldflags = CGO_LDFLAGS="$(CURDIR)/src/common/erasure/longhair/bin/liblonghair.a -lstdc++"
govars = $(gopath) $(cgo_ldflags)
packages = common network quorum disk main common/erasure common/crypto common/log

all: submodule-update fmt libraries

submodule-update:
	git submodule update

fmt:
	$(govars) go fmt $(packages)

libraries:
	$(govars) go install $(packages)

test: libraries
	$(govars) go test $(packages)

race-libs:
	$(govars) go install -race std

bench: libraries
	$(govars) go test $(packages)

test-verbose: libraries
	$(govars) go test -test.v $(packages)

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all test fmt libraries test-verbose docs submodule-update
