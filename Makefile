# SUBDIRS := data mesh messanger otto server station utils
PIENV	= env GOOS=linux GOARCH=arm GOARM=7

all: test $(SUBDIRS) build

init:
	git update --init 

build:
	rm -f garden-station
	go build -v . 

fmt:
	gofmt -w .

test:
	rm -f cover.out
	go test -coverprofile=cover.out -cover ./...

verbose:
	rm -f cover.out
	go test -v -coverprofile=cover.out -cover ./...

coverage: test
	go tool cover -func=cover.out

html: test
	rm -f coverage.html
	go tool cover -html=cover.out -o coverage.html

.PHONY: all test build fmt $(SUBDIRS)
