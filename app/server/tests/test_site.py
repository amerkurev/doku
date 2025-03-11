import time
from pathlib import Path
from unittest.mock import patch, MagicMock

import pytest
from fastapi.testclient import TestClient

import settings
from main import app
from contrib.types import (
    DockerImage,
    DockerImageList,
    DockerContainer,
    DockerContainerList,
    DockerVolume,
    DockerVolumeList,
    DockerBuildCache,
    DockerBuildCacheList,
    DockerMount,
    DockerContainerLog,
    DockerBindMounts,
    DockerOverlay2Layer,
)


client = TestClient(app)


@pytest.fixture
def log_start_time():
    def _log_start_time(database: Path, table: str) -> int:
        ts = int(time.time())
        filename = database / f'{table}.timestamp'
        with filename.open('w') as fd:
            fd.write(str(ts))
        return ts

    return _log_start_time


@patch('server.router.context.kvstore')
def test_images(mock_kvstore, log_start_time):
    settings.DB_DF.touch()
    log_start_time(settings.DB_DIR, settings.TABLE_SYSTEM_DF)

    images = [
        DockerImage(
            Id='sha256:123456789abcdef',
            Created='2023-01-01T12:00:00Z',
            ParentID='sha256:987654321fedcba',
            RepoTags=['image:latest', 'image:v1'],
            SharedSize=1000,
            Size=4000,
            containers=['test_container1', 'test_container2'],
        ),
    ]
    mock_kvstore.get.return_value = MagicMock(root=images)

    response = client.get('/site/images/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'


@patch('server.router.context.kvstore')
def test_containers(mock_kvstore, log_start_time):
    settings.DB_DF.touch()
    log_start_time(settings.DB_DIR, settings.TABLE_SYSTEM_DF)

    containers = [
        DockerContainer(
            Id='sha256:123456789abcdef',
            Names=['/test_container1', '/test_container2'],
            Image='image:latest',
            ImageID='sha256:987654321fedcba',
            Created='2023-01-01T12:00:00Z',
            SizeRw=50000,
            SizeRootFs=200000,
            State='running',
        ),
    ]
    mock_kvstore.get.return_value = MagicMock(root=containers)

    response = client.get('/site/containers/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'


@patch('server.router.context.kvstore')
def test_volumes(mock_kvstore, log_start_time):
    settings.DB_DF.touch()
    log_start_time(settings.DB_DIR, settings.TABLE_SYSTEM_DF)

    volumes = [
        DockerVolume(
            Name='test_volume1',
            Driver='local',
            CreatedAt='2023-01-01T12:00:00Z',
            Mountpoint='/var/lib/docker/volumes/test_volume1/_data',
            Scope='local',
            UsageData={'Size': 1000, 'RefCount': 2},
        ),
    ]
    mock_kvstore.get.return_value = MagicMock(root=volumes)

    response = client.get('/site/volumes/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'


@patch('server.router.context.kvstore')
def test_build_cache(mock_kvstore, log_start_time):
    settings.DB_DF.touch()
    log_start_time(settings.DB_DIR, settings.TABLE_SYSTEM_DF)

    build_cache = [
        DockerBuildCache(
            ID='sha256:123456789abcdef',
            Type='layer',
            Description='mount / from exec /bin/sh -c apk update && apk add postgresql-dev',
            InUse=True,
            Shared=True,
            Size=2048,
            CreatedAt='2023-01-01T12:00:00Z',
            LastUsedAt='2023-01-01T12:00:00Z',
            UsageCount=2,
        ),
    ]
    mock_kvstore.get.return_value = MagicMock(root=build_cache)

    response = client.get('/site/build-cache/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'


@patch('server.router.context.kvstore')
def test_bind_mounts(mock_kvstore, log_start_time):
    settings.DB_DU.touch()
    log_start_time(settings.DB_DIR, settings.TABLE_BINDMOUNTS)

    bind_mounts = [
        DockerBindMounts(
            path='/mnt/bind',
            err=False,
            size=100000,
            scan_in_progress=True,
            last_scan='2023-01-01T12:00:00Z',
            containers=['container1', 'container2'],
        ),
    ]
    mock_kvstore.get_all.return_value = bind_mounts

    response = client.get('/site/bind-mounts/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'


@patch('server.router.context.kvstore')
def test_logs(mock_kvstore, log_start_time):
    settings.DB_DF.touch()
    log_start_time(settings.DB_DIR, settings.TABLE_LOGFILES)

    logs = [
        DockerContainerLog(
            id='123456789abc',
            name='nginx',
            image='nginx:latest',
            path='/var/lib/docker/containers/7d2de847bebae847b-json.log',
            size=50000,
            last_scan='2023-01-01T12:00:00Z',
        ),
    ]
    mock_kvstore.get_all.return_value = logs

    response = client.get('/site/logs/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'


@patch('server.router.context.kvstore')
def test_overlay2(mock_kvstore, log_start_time):
    settings.DB_DU.touch()
    log_start_time(settings.DB_DIR, settings.TABLE_OVERLAY2)

    overlay2 = [
        DockerOverlay2Layer(
            id='jq1br13jcumomv9j6u8rce531485cee0f83624769a2d',
            created='2023-01-01T12:00:00Z',
            diff_root='/usr, /etc, /opt, /var, /entrypoint.sh, /sys, /sbin, /mnt, /bitnami, /run, /lib64, /boot, /proc, /dev, /media, /lib, /srv, /tmp, /home, /bin, /root, /run.sh',
            err=False,
            size=5000,
            scan_in_progress=False,
            last_scan='2023-02-01T12:00:00Z',
            in_use=True,
        ),
    ]
    mock_kvstore.get_all.return_value = overlay2

    response = client.get('/site/overlay2/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'


@patch('docker.api.client.APIClient.version')
@patch('server.router.context.kvstore')
@patch('contrib.docker.docker_from_env')
def test_dashboard(
    mock_docker_from_env,
    mock_kvstore,
    mock_version,
):
    settings.DB_DF.touch()
    settings.DB_DU.touch()

    version_data = {
        'Platform': {'Name': 'Docker Engine - Community'},
        'Version': '20.10.7',
        'ApiVersion': '1.41',
        'MinAPIVersion': '1.12',
        'Os': 'linux',
        'Arch': 'amd64',
        'KernelVersion': '5.10.25-linuxkit',
    }

    mock_version.return_value = version_data

    mock_client = MagicMock()
    mock_client.version.return_value = version_data
    mock_docker_from_env.return_value = mock_client

    # Define mock data for different keys
    mock_data = {
        DockerImageList: [
            DockerImage(
                Id='sha256:123456789abcdef',
                Created='2023-01-01T12:00:00Z',
                ParentID='sha256:987654321fedcba',
                RepoTags=['image:latest'],
                SharedSize=1000,
                Size=4000,
                containers=[],
            )
        ],
        DockerContainerList: [
            DockerContainer(
                Id='sha256:123456789abcdef',
                Names=['/test_container'],
                Image='image:latest',
                ImageID='sha256:987654321fedcba',
                Created='2023-01-01T12:00:00Z',
                SizeRw=50000,
                SizeRootFs=200000,
                State='running',
            )
        ],
        DockerVolumeList: [
            DockerVolume(
                Name='test_volume',
                Driver='local',
                CreatedAt='2023-01-01T12:00:00Z',
                Mountpoint='/var/lib/docker/volumes/test_volume/_data',
                Scope='local',
                UsageData={'Size': 1000, 'RefCount': 2},
            )
        ],
        DockerBuildCacheList: [
            DockerBuildCache(
                ID='sha256:123456789abcdef',
                Type='layer',
                Description='mount / from exec /bin/sh -c apk update',
                InUse=True,
                Shared=True,
                Size=2048,
                CreatedAt='2023-01-01T12:00:00Z',
                LastUsedAt='2023-01-01T12:00:00Z',
                UsageCount=2,
            )
        ],
        DockerBindMounts: [
            DockerBindMounts(
                path='/mnt/bind',
                err=False,
                size=100000,
                scan_in_progress=False,
                last_scan='2023-01-01T12:00:00Z',
                containers=['container1'],
            )
        ],
        DockerContainerLog: [
            DockerContainerLog(
                id='123456789abc',
                name='nginx',
                image='nginx:latest',
                path='/var/lib/docker/containers/123456789abc-json.log',
                size=50000,
                last_scan='2023-01-01T12:00:00Z',
            )
        ],
        DockerOverlay2Layer: [
            DockerOverlay2Layer(
                id='jq1br13jcumomv9j6u8rce531',
                created='2023-01-01T12:00:00Z',
                diff_root='/var/lib/docker/overlay2/diff/jq1br13jcumomv9j6u8rce531',
                err=False,
                size=5000,
                scan_in_progress=False,
                last_scan='2023-01-01T12:00:00Z',
                in_use=True,
            )
        ],
        DockerMount: [
            DockerMount(
                Source='/host/path',
                Destination='/container/path',
                Mode='ro',
                Propagation='rprivate',
                RW=False,
                Type='bind',
                Root=True,
            )
        ],
    }

    def get_side_effect(key, kv, model):
        return MagicMock(root=mock_data[model])

    def get_all_side_effect(kv, model):
        return mock_data[model]

    mock_kvstore.get.side_effect = get_side_effect
    mock_kvstore.get_all.side_effect = get_all_side_effect

    response = client.get('/site/')
    assert response.status_code == 200
    assert response.headers['content-type'] == 'text/html; charset=utf-8'

    assert mock_version.called
