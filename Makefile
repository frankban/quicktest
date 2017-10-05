default: test

$(GOPATH)/bin/godeps:
	go get -v github.com/rogpeppe/godeps

deps: $(GOPATH)/bin/godeps
	$(GOPATH)/bin/godeps -u dependencies.tsv

create-deps: $(GOPATH)/bin/godeps
	$(GOPATH)/bin/godeps -t ./... > dependencies.tsv

build: deps
	go build -v ./...

install: deps
	go install -v ./...

test: deps
	go test -v ./...
