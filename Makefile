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

test-verbose: libraries
	$(govars) go test -test.v $(packages)

bench: libraries
	$(govars) go test $(packages)

dependencies:
	cd src/common/crypto/libsodium && ./autogen.sh && ./configure && make check && sudo make install && sudo ldconfig

race-libs:
	$(govars) go install -race std

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all submodule-update fmt libraries test test-verbose bench dependencies race-libs docs
