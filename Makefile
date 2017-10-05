default: test

$(GOPATH)/bin/godeps:
	go get -v github.com/rogpeppe/godeps

deps: $(GOPATH)/bin/godeps
	$(GOPATH)/bin/godeps -u dependencies.tsv

create-deps: $(GOPATH)/bin/godeps
	$(GOPATH)/bin/godeps -t ./... > dependencies.tsv

.PHONY: build
build: deps
	go build -v ./...

.PHONY: install
install: deps
	go install -v ./...

.PHONY: test
test: deps
	go test -v ./...
