from pathlib import Path
from unittest.mock import Mock, patch

import docker
import pytest

from contrib.docker import (
    docker_from_env,
    doku_container,
    doku_mounts,
    _get_mounts,
    map_host_path_to_container,
)


@pytest.fixture
def docker_client_mock():
    return Mock(spec=docker.DockerClient)


@pytest.fixture
def container_mock():
    container = Mock()
    container.attrs = {
        'Config': {'Hostname': 'test-hostname'},
        'Mounts': [
            {
                'Type': 'bind',
                'Source': '/host/path',
                'Destination': '/container/path',
                'Mode': 'rw',
                'RW': True,
                'Propagation': 'rprivate',
            },
            {
                'Type': 'bind',
                'Source': '/var/run/docker.sock',
                'Destination': '/var/run/docker.sock',
                'Mode': 'rw',
                'RW': True,
                'Propagation': 'rprivate',
            },
        ],
        'LogPath': '/host/path/logs',
    }
    return container


@patch('contrib.docker.docker.from_env')
@patch('contrib.docker.settings')
def test_docker_from_env(settings_mock, docker_from_env_mock):
    settings_mock.DOCKER_VERSION = '1.41'
    settings_mock.DOCKER_TIMEOUT = 30
    settings_mock.DOCKER_MAX_POOL_SIZE = 10
    settings_mock.DOCKER_ENV = {'foo': 'bar'}
    settings_mock.DOCKER_USE_SSH_CLIENT = False

    docker_from_env()
    docker_from_env_mock.assert_called_once_with(
        version='1.41',
        timeout=30,
        max_pool_size=10,
        environment={'foo': 'bar'},
        use_ssh_client=False,
    )


@patch('contrib.docker.settings')
def test_doku_container_not_in_docker(settings_mock, docker_client_mock):
    settings_mock.IN_DOCKER = False
    assert doku_container(docker_client_mock) is None


@patch('contrib.docker.settings')
def test_doku_container_match_hostname(settings_mock, docker_client_mock, container_mock):
    settings_mock.IN_DOCKER = True
    settings_mock.GITHUB_REPO = 'test/repo'
    settings_mock.MY_HOSTNAME = 'test-hostname'

    container2 = Mock()
    container2.attrs = {'Config': {'Hostname': 'other-hostname'}}

    docker_client_mock.containers.list.return_value = [container2, container_mock]
    assert doku_container(docker_client_mock) == container_mock


def test_doku_mounts(docker_client_mock, container_mock):
    docker_client_mock.containers.list.return_value = [container_mock]
    mounts = doku_mounts(docker_client_mock)
    assert len(mounts) == 1
    assert mounts[0].src == '/host/path'
    assert mounts[0].dst == '/container/path'
    assert mounts[0].mode == 'rw'


def test_get_mounts(container_mock):
    mounts = _get_mounts(container_mock)
    assert len(mounts) == 1
    assert mounts[0].dst == '/container/path'
    assert mounts[0].root is False


@patch('pathlib.Path.exists')
def test_map_host_path_to_container(mock_exists):
    mock_exists.return_value = True
    result = map_host_path_to_container('/source', '/dest', '/source/subdir/file.txt')
    assert result == Path('/dest/subdir/file.txt')


@patch('pathlib.Path.exists')
def test_map_host_path_host_mnt_prefix(mock_exists):
    mock_exists.return_value = True
    result = map_host_path_to_container('/host_mnt/source', '/dest', '/source/file.txt')
    assert result == Path('/dest/file.txt')
