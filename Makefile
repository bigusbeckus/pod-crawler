OUTFILE = bin/podcrawler

all: staticcheck build

fmtcheck:
	"$(CURDIR)/scripts/gofmtcheck.sh"

staticcheck:
	"$(CURDIR)/scripts/staticcheck.sh"

build:
	"$(CURDIR)/scripts/build.sh" -o $(OUTFILE)

clean:
	"$(CURDIR)/scripts/build.sh" -c

run:
	./$(OUTFILE)

test:
	go test -v ./...

debug: build run

.PHONY: fmtcheck staticcheck
