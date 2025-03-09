GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GIT_SHA=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
BUILD_DATE=$(shell date +%Y%m%d-%H:%M:%S)
# If we're exactly on a tag, just show tag, otherwise show detailed version with commits
GIT_DESCRIBE=$(shell git describe --tags 2>/dev/null)
ifeq ($(GIT_TAG),$(GIT_DESCRIBE))
    # We're exactly on a tag
    REV=$(GIT_TAG) ($(BUILD_DATE))
else
    # We're off a tag
    REV=$(shell git describe --tags) ($(BUILD_DATE))
endif
PWD=$(shell pwd)

info:
	- @echo "revision $(REV)"

build:
	- @docker buildx build --build-arg REV="${REV}" -t amerkurev/doku:latest --progress=plain .

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
