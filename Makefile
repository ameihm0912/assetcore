export GOPATH = $(shell pwd)
PROJ = assetcore

all: $(PROJ)

depends:
	go get github.com/bitly/go-hostpool
	go get github.com/araddon/gou
	go get code.google.com/p/go-uuid/uuid
	go get github.com/mattbaird/elastigo

cleanall:
	rm -rf pkg
	rm -rf src/github.com
	rm -rf src/code.google.com
	rm -rf bin/*

clean:
	rm -f bin/assetcore

assetcore:
	go install assetcore
