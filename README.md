# Doku
Doku is a simple, lightweight web-based application that allows you to monitor Docker disk usage in a user-friendly manner.
The Doku displays the amount of disk space used by the Docker daemon, splits by images, containers, volumes, and builder cache.
If you're lucky, you'll also see the sizes of log files :)

Doku should work for most. It has been tested with dozen of hosts.
<div>

[![Build](https://github.com/amerkurev/doku/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/amerkurev/doku/actions/workflows/ci.yml)&nbsp;
[![Coverage Status](https://coveralls.io/repos/github/amerkurev/doku/badge.svg?branch=master)](https://coveralls.io/github/amerkurev/doku?branch=master)&nbsp;
[![GoReportCard](https://goreportcard.com/badge/github.com/amerkurev/doku)](https://goreportcard.com/report/github.com/amerkurev/doku)&nbsp;
[![Docker Hub](https://img.shields.io/docker/automated/amerkurev/doku.svg)](https://hub.docker.com/r/amerkurev/doku/tags)&nbsp;
[![Docker pulls](https://img.shields.io/docker/pulls/amerkurev/doku.svg)](https://hub.docker.com/r/amerkurev/doku)&nbsp;
</div>

![laptop_doku](https://user-images.githubusercontent.com/28217522/235870076-a344527c-874d-41a4-bda9-749efd4ff917.svg)

## Getting Doku

Doku is a very small Docker container (6 MB compressed). Pull the latest release from the index:

    docker pull amerkurev/doku:latest
    
## Using Doku

The simplest way to use Doku is to run the Docker container. Mount the Docker Unix socket with `-v` to `/var/run/docker.sock`. Also, you need to mount the top-level directory (`/`) on the host machine in `ro` mode. Otherwise, Doku will not be able to calculate the size of the logs and bind mounts.

    docker run --name doku -d -v /var/run/docker.sock:/var/run/docker.sock:ro -v /:/hostroot:ro -p 9090:9090 amerkurev/doku

Doku will be available at [http://localhost:9090/](http://localhost:9090/). You can change `-p 9090:9090` to any port. For example, if you want to view Doku over port 8080 then you would do `-p 8080:9090`.

## Basic auth

Doku supports basic auth for all requests. This functionality is disabled by default.

In order to enable basic auth, user should set the typical htpasswd file with `--basic-htpasswd=<file location>` or `env BASIC_HTPASSWD=<file location>`. 

Doku expects htpasswd file to be in the following format:

```
username1:bcrypt(password2)
username2:bcrypt(password2)
...
```

this can be generated with `htpasswd -nbB` command, i.e. `htpasswd -nbB test passwd`

## Supported architectures
- linux/amd64
- linux/arm/v7
- linux/arm64

## Special thanks to

The following great works inspired me:
- Gatus: https://github.com/TwiN/gatus
- Dozzle: https://github.com/amir20/dozzle
- Reproxy: https://github.com/umputun/reproxy

## License

[MIT](LICENSE)
