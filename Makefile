ifdef GITHUB_REF
    # CI environment - use provided ref
    GIT_TAG=$(shell echo $(GITHUB_REF) | cut -d'/' -f3)
else
    # Local environment - calculate from git
	GIT_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || git rev-parse --abbrev-ref HEAD)
endif

ifdef GITHUB_SHA
	# CI environment - use provided sha
	GIT_SHA=$(GITHUB_SHA)
else
	# Local environment - calculate from git
	GIT_SHA=$(shell git rev-parse --short HEAD)
endif

REV=$(GIT_TAG)-$(GIT_SHA)
PWD=$(shell pwd)

info:
	- @echo "revision $(REV)"

build:
	- @docker buildx build --load --build-arg GIT_SHA="${GIT_SHA}" --build-arg GIT_TAG="${GIT_TAG}" -t amerkurev/doku:latest --progress=plain .

lint:
	- @ruff check app

fmt: lint
	- @ruff format app

test:
	- @docker run --rm -t -v $(PWD)/app:/usr/src/app amerkurev/doku ./pytest.sh

cov:
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

.PHONY: info build lint fmt test cov dev
