# Make is only required when using Go < 1.11.
# On newer Go versions dependency management is handled by Go modules.

default: test

$(GOPATH)/bin/godeps:
	go get -v github.com/rogpeppe/godeps

deps: $(GOPATH)/bin/godeps
	$(GOPATH)/bin/godeps -u dependencies.tsv

create-deps: $(GOPATH)/bin/godeps
	$(GOPATH)/bin/godeps -t ./... > dependencies.tsv

.PHONY: install
install: deps
	go install -v ./...

.PHONY: test
test: deps
	go test -v ./...
