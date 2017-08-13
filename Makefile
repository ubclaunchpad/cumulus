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
	glide install

clean: cumulus
	rm cumulus

install-glide:
	sh scripts/install_glide.sh
