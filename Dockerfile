FROM golang:1.18-alpine as backend

ARG GIT_BRANCH
ARG GITHUB_SHA
ARG CI

ENV GOFLAGS="-mod=vendor"
ENV CGO_ENABLED=0
ENV GOOS=linux

ADD . /build
WORKDIR /build

RUN apk add --no-cache --update git tzdata ca-certificates

RUN \
    if [ -z "$CI" ] ; then \
    echo "runs outside of CI" && version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format=%h)-$(date +%Y%m%dT%H:%M:%S); \
    else version=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H:%M:%S); fi && \
    echo "version=$version" && \
    cd app && go build -o /build/doku -ldflags "-X main.revision=${version} -s -w"


FROM node:16.14.0-alpine as build-frontend

ARG NODE_ENV=production
ARG CI=true

ADD web/doku /srv/frontend
WORKDIR /srv/frontend

RUN rm -f /srv/frontend/.eslintrc.json && \
    apk update && \
    apk add zip make gcc g++ python3 && \
    yarn install --immutable && \
    yarn semantic-ui-css-patch && \
    yarn build
CMD yarn run test


FROM ghcr.io/umputun/baseimage/app:v1.9.1 as base

FROM scratch

ENV DOKU_IN_DOCKER=1

COPY --from=backend /build/doku /srv/doku
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group
COPY --from=build-frontend /srv/frontend/build/static /srv/web/static
COPY --from=build-frontend /srv/frontend/build/favicon.ico /srv/web/static
COPY --from=build-frontend /srv/frontend/build/index.html /srv/web/static
COPY --from=build-frontend /srv/frontend/build/manifest.json /srv/web/static

WORKDIR /srv
ENTRYPOINT ["/srv/doku"]
