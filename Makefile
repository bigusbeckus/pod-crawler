all: staticcheck build

fmtcheck:
	"$(CURDIR)/scripts/gofmtcheck.sh"

staticcheck:
	"$(CURDIR)/scripts/staticcheck.sh"

build:
	"$(CURDIR)/scripts/build.sh"

clean:
	"$(CURDIR)/scripts/build.sh" -c

debug:
	go run .

.PHONY: fmtcheck staticcheck
