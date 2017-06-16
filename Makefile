.PHONY: test deps clean install_glide

PACKAGES = `go list ./... | grep -v vendor/`

all: cumulus

cumulus:
	go build

test:
	go test $(PACKAGES)

deps:
	glide install

clean:
	rm cumulus

install-glide:
	sh scripts/install_glide.sh
