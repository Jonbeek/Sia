gopath = GOPATH=$(CURDIR)
cgo_ldflags = CGO_LDFLAGS="$(CURDIR)/src/common/erasure/longhair/bin/liblonghair.a -lstdc++"
govars = $(gopath) $(cgo_ldflags)
packages = common common/crypto common/erasure common/log disk main network quorum test

all: submodule-update libraries

submodule-update:
	git submodule update --init

fmt:
	$(govars) go fmt $(packages)

libraries: fmt
	$(govars) go install $(packages)

test: libraries
	$(govars) go test -short $(packages)

test-verbose: libraries
	$(govars) go test -short -v $(packages)

test-race: libraries
	$(govars) go test -short -race $(packages)

test-race-verbose: libraries
	$(govars) go test -short -race -v $(packages)

test-long: libraries
	$(govars) go test -race $(packages)

test-long-verbose: libraries
	$(govars) go test -v -race $(packages)

dependencies: submodule-update race-libs
	cd src/common/crypto/libsodium && ./autogen.sh && ./configure && make check && sudo make install && sudo ldconfig

race-libs:
	$(govars) go install -race std

docs:
	pdflatex -output-directory=doc/ doc/whitepaper.tex 

.PHONY: all submodule-update fmt libraries test test-long dependencies race-libs docs
