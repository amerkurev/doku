from pathlib import Path

import docker
from docker.models.containers import Container
from pydantic import ValidationError

import settings
from contrib.types import DockerMount


def docker_from_env() -> docker.DockerClient:
    """
    Create a Docker client with settings from environment.
    """
    client = docker.from_env(
        version=settings.DOCKER_VERSION,
        timeout=settings.DOCKER_TIMEOUT,
        max_pool_size=settings.DOCKER_MAX_POOL_SIZE,
        environment=settings.DOCKER_ENV,
        use_ssh_client=settings.DOCKER_USE_SSH_CLIENT,
    )
    return client


def doku_container(client: docker.DockerClient) -> Container | None:
    """
    Get the Doku container (current container).
    """
    if not settings.IN_DOCKER:
        return None

    # filter by label github.repo=settings.GITHUB_REPO
    filters = {'label': 'github.repo=' + settings.GITHUB_REPO}
    c = client.containers.list(filters=filters)
    if not c:
        return None

    # if only one container is found, return it
    if len(c) == 1:
        return c[0]

    # if multiple containers are found, return the container with the same hostname
    if not settings.MY_HOSTNAME:
        return None

    for cont in c:
        attrs = cont.attrs
        if 'Config' in attrs and 'Hostname' in attrs['Config']:
            hostname = attrs['Config']['Hostname']
            if hostname == settings.MY_HOSTNAME:
                return cont
    return None


def doku_mounts(client: docker.DockerClient) -> list[DockerMount]:
    """
    Get all mounts of the Doku container.
    """
    cont = doku_container(client)
    return _get_mounts(cont) if cont else []


def _get_mounts(cont: Container) -> list[DockerMount]:
    if 'Mounts' not in cont.attrs:
        return []

    ret = []
    mounts = cont.attrs['Mounts']
    log_path = cont.attrs.get('LogPath')

    for mount in mounts:
        try:
            mnt = DockerMount.model_validate(mount)
        except ValidationError:
            continue

        # skip mounts to docker socket
        if mnt.dst == '/var/run/docker.sock':
            continue

        # check if the mount is a root mount
        if log_path and map_host_path_to_container(mnt.src, mnt.dst, log_path):
            mnt.root = True

        ret.append(mnt)

    return ret


def map_host_path_to_container(source: str, destination: str, host_path: str) -> Path | None:
    """
    Map host file path to container path based on mount configuration.

    Args:
        source: Mount source path on host
        destination: Mount destination path in container
        host_path: Path on host to map
    """
    # remove trailing slashes for consistency
    source = source.rstrip('/')
    destination = destination.rstrip('/')
    host_path = host_path.rstrip('/')

    # check if host path is under mount source
    if not host_path.startswith(source):
        # if source is /host_mnt, try to remove it and check again
        if source.startswith('/host_mnt'):
            source = source.removeprefix('/host_mnt')
            return map_host_path_to_container(source, destination, host_path)

        return None

    # remove source prefix and join with destination
    relative_path = host_path[len(source) :]
    path = Path(destination + relative_path)
    return path if path.exists() else None
