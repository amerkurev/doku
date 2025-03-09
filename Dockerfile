FROM python:3.13-slim AS base

FROM base AS builder

RUN apt update && apt install -y python3-dev && mkdir /install
WORKDIR /install
COPY requirements.txt ./
RUN pip install --no-cache-dir --prefix=/install -r ./requirements.txt


FROM base AS final

COPY --from=builder /install /usr/local


ARG APP_DIR=/usr/src/app
ARG SUPERVISOR_CONF=/usr/local/etc/supervisord.conf
ARG GITHUB_REPO=amerkurev/doku
ARG GIT_SHA
ARG GIT_TAG

# this is used in the code
LABEL github.repo=$GITHUB_REPO
LABEL org.opencontainers.image.title="Doku"
LABEL org.opencontainers.image.description="Doku - Docker disk usage dashboard"
LABEL org.opencontainers.image.url="https://docker-disk.space"
LABEL org.opencontainers.image.documentation="https://github.com/amerkurev/doku?tab=readme-ov-file#doku"
LABEL org.opencontainers.image.vendor="amerkurev"
LABEL org.opencontainers.image.licenses="MIT License"
LABEL org.opencontainers.image.source="https://github.com/amerkurev/doku"

ENV IN_DOCKER=1 \
	# used in supervisord.conf
	APP_DIR=$APP_DIR \
	GITHUB_REPO=$GITHUB_REPO \
	GIT_SHA=$GIT_SHA \
	GIT_TAG=$GIT_TAG


COPY app $APP_DIR
COPY conf/supervisord.conf $SUPERVISOR_CONF

WORKDIR $APP_DIR

CMD ["supervisord", "-c", "/usr/local/etc/supervisord.conf"]
