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

out:
	echo "$(OUTFILE)"

test:
	go test -v ./...

debug: build run

crawl:
	docker compose --env-file ./.env -f deploy/docker-compose.yml up --build

crawld:
	docker compose --env-file ./.env -f deploy/docker-compose.yml up -d --build

.PHONY: *
