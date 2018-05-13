.PHONY: test deps clean install_glide

all: cumulus

cumulus:
	go build

test:
	go test ./... --cover

run-console: cumulus
	./cumulus run -c

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

clean: cumulus
	rm -f cumulus blockchain.json user.json logfile
