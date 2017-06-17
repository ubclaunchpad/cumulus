PACKAGES = `go list ./... | grep -v vendor/`

all: cumulus

cumulus:
	go build

test:
	go test $(PACKAGES) --cover

deps:
	glide install

clean:
	rm cumulus

install-glide:
	sh scripts/install_glide.sh
