export GOPATH = $(shell pwd)
PROJ = assetcore

all: $(PROJ)

predeps:
	go get github.com/bitly/go-hostpool
	go get github.com/araddon/gou

depends: predeps setup-elastigo build-elastigo

build-elastigo:
	go install github.com/mattbaird/elastigo

setup-elastigo:
	mkdir -p src/github.com/mattbaird
	git clone https://github.com/mattbaird/elastigo.git \
		src/github.com/mattbaird/elastigo
	(cd src/github.com/mattbaird/elastigo && \
		git checkout tags/v1.0)

cleanall:
	rm -rf pkg
	rm -rf src/github.com
	rm -rf bin/*

clean:
	rm -f bin/assetcore

assetcore:
	go install assetcore
