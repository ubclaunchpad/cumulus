.PHONY: test deps clean install_glide

PACKAGES = `go list ./... | grep -v vendor/`

all: cumulus

cumulus:
	go build

test:
	go test $(PACKAGES) --cover

run-console: cumulus
	./cumulus run -c

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

clean: cumulus
	rm -f cumulus blockchain.json user.json logfile
