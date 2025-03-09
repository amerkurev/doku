BUILD_DATE=$(shell date +%Y%m%d-%H:%M:%S)

ifdef GITHUB_REF
    # CI environment - use provided ref
    REF=$(shell echo $(GITHUB_REF) | cut -d'/' -f3)
else
    # Local environment - calculate from git
    REF=$(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
endif

REV=$(REF) ($(BUILD_DATE))
PWD=$(shell pwd)

info:
	- @echo "revision $(REV)"

build:
	- @docker buildx build --load --build-arg REV="${REV}" -t amerkurev/doku:latest --progress=plain .

lint:
	- @ruff check app

fmt: lint
	- @ruff format app

test:
	- @docker run --rm -t -v $(PWD)/app:/usr/src/app amerkurev/doku ./pytest.sh
	- @docker run --rm -t -v $(PWD)/app:/usr/src/app amerkurev/doku coverage html

dev: build
	- @docker run \
	  -it \
	  --rm \
	  --name doku \
  	  --env-file=.env \
	  -v $(PWD)/app:/usr/src/app \
	  -v /var/run/docker.sock:/var/run/docker.sock:ro \
	  -v /:/hostroot:ro \
      -v $(PWD)/.htpasswd:/.htpasswd \
	  -v $(PWD)/.ssl/key.pem:/.ssl/key.pem \
	  -v $(PWD)/.ssl/cert.pem:/.ssl/cert.pem \
	  -p 9090:9090 \
	  amerkurev/doku

.PHONY: info build lint fmt dev
