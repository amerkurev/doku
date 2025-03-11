from pathlib import Path
from unittest.mock import MagicMock, patch, call, ANY

import pytest
from docker.models.images import Image
from docker.models.containers import Container

from contrib.types import (
    DockerMount,
    DockerImageList,
    DockerContainerList,
    DockerVolumeList,
    DockerBuildCacheList,
    DockerContainerLog,
    DockerBindMounts,
    DockerOverlay2Layer,
)
from scan.scanner import BaseScanner, SystemDFScanner, LogfilesScanner, BindMountsScanner, Overlay2Scanner


@pytest.fixture
def mock_docker_client():
    return MagicMock()


@pytest.fixture
def mock_is_stop():
    return MagicMock(return_value=False)


@pytest.fixture
def docker_mount():
    return DockerMount.model_validate({
        'Source': '/host/path',
        'Destination': '/container/path',
        'Mode': 'ro',
        'Propagation': 'rprivate',
        'RW': False,
        'Type': 'bind',
        'Root': True,
    })


def test_base_scanner(mock_docker_client):
    with patch('scan.scanner.docker_from_env', return_value=mock_docker_client):
        scanner = BaseScanner()
        assert scanner.logger is not None

        with pytest.raises(NotImplementedError):
            scanner.scan()

        with pytest.raises(NotImplementedError):
            _ = scanner.database_name

        with pytest.raises(NotImplementedError):
            _ = scanner.table_name


def test_system_df_scanner(mock_docker_client, docker_mount):
    # mock DF data
    df_data = {
        'Images': [
            {
                'Id': 'sha256:123456789abcdef',
                'Created': '2023-01-01T12:00:00Z',
                'ParentId': 'sha256:987654321fedcba',
                'RepoTags': ['image:latest', 'image:v1'],
                'SharedSize': 1000,
                'Size': 5000,
            }
        ],
        'Containers': [
            {
                'Id': '123456789abcdef',
                'Names': ['/container_name'],
                'Image': 'nginx:latest',
                'ImageID': 'sha256:123456789abcdef',
                'Created': '2023-01-01T12:00:00Z',
                'SizeRw': 50000,
                'SizeRootFs': 200000,
                'State': 'running',
            }
        ],
        'Volumes': [
            {
                'Name': 'volume1',
                'Driver': 'local',
                'CreatedAt': '2023-01-01T12:00:00Z',
                'Mountpoint': '/var/lib/docker/volumes/volume1',
                'Scope': 'local',
                'UsageData': {'Size': 1000, 'RefCount': 2},
            }
        ],
        'BuildCache': [
            {
                'ID': 'abcdef123456',
                'Type': 'layer',
                'Description': 'mount / from exec /bin/sh -c apk update && apk add ' * 10,
                'InUse': True,
                'Shared': False,
                'Size': 1024,
                'CreatedAt': '2023-01-01T12:00:00Z',
                'LastUsedAt': '2023-02-01T12:00:00Z',
                'UsageCount': 5,
            }
        ],
    }

    # container with image and volume
    mock_container = MagicMock(spec=Container)
    mock_container.name = 'container1'
    mock_container.attrs = {'Image': 'sha256:123456789abcdef', 'Mounts': [{'Type': 'volume', 'Name': 'volume1'}]}

    mock_docker_client.df.return_value = df_data
    mock_docker_client.containers.list.return_value = [mock_container]

    with (
        patch('scan.scanner.docker_from_env', return_value=mock_docker_client),
        patch('scan.scanner.doku_mounts', return_value=[docker_mount]),
        patch('scan.scanner.kvstore.set') as mock_kvstore_set,
    ):
        scanner = SystemDFScanner()
        scanner.client = mock_docker_client
        scanner.scan()

        mock_docker_client.df.assert_called_once()
        mock_docker_client.containers.list.assert_called_once_with(all=True, sparse=False)

        images = DockerImageList.model_validate(df_data['Images'])
        images.root[0].containers = ['container1']

        volumes = DockerVolumeList.model_validate(df_data['Volumes'])
        volumes.root[0].containers = ['container1']

        # check what was stored in the kvstore
        mock_kvstore_set.assert_has_calls(
            [
                call('image', images, ANY),
                call('container', DockerContainerList.model_validate(df_data['Containers']), ANY),
                call('volume', volumes, ANY),
                call('build_cache', DockerBuildCacheList.model_validate(df_data['BuildCache']), ANY),
                call('root_mount', docker_mount, ANY),
            ],
            any_order=True,
        )


def test_logfiles_scanner(mock_docker_client, mock_is_stop, docker_mount):
    short_id = '7d2de847bebae847b'
    name = 'container1'
    log_path = '/var/lib/docker/containers/7d2de847bebae847b-json.log'
    image = 'nginx:latest'
    st_size = 1024 * 10

    # mock container data
    mock_container = MagicMock(spec=Container)
    mock_container.short_id = short_id
    mock_container.name = name
    mock_container.attrs = {'LogPath': log_path}

    mock_image = MagicMock(spec=Image)
    mock_image.tags = [image]
    mock_container.image = mock_image

    mock_docker_client.containers.list.return_value = [mock_container]

    with (
        patch('scan.scanner.docker_from_env', return_value=mock_docker_client),
        patch('scan.scanner.doku_mounts', return_value=[docker_mount]),
        patch('scan.scanner.map_host_path_to_container') as mock_map_path,
        patch('scan.scanner.kvstore.set') as mock_kvstore_set,
    ):
        mock_path = MagicMock(spec=Path)
        mock_path.stat.return_value.st_size = st_size
        mock_map_path.return_value = mock_path

        scanner = LogfilesScanner(mock_is_stop)
        scanner.client = mock_docker_client
        scanner.scan()

        mock_docker_client.containers.list.assert_called_once_with(all=True)

        # check what was stored in the kvstore
        obj = DockerContainerLog.model_validate({
            'id': short_id,
            'name': name,
            'image': image,
            'path': log_path,
            'size': st_size,
            'last_scan': '2023-01-01T12:00:00Z',
        })
        obj.last_scan = ANY
        mock_kvstore_set.assert_called_once_with(mock_container.short_id, obj, ANY)

        mock_map_path.return_value = None
        scanner.scan()


def test_bind_mounts_scanner(mock_docker_client, mock_is_stop, docker_mount):
    # mock Doku container
    doku_container = MagicMock(spec=Container)
    doku_container.short_id = 'doku1234'
    doku_container.name = 'doku_container'
    doku_container.attrs = {
        'Mounts': [
            {'Type': 'bind', 'Source': '/doku/path', 'Destination': '/doku/container/path', 'Mode': 'rw', 'RW': True}
        ]
    }
    doku_image = MagicMock(spec=Image)
    doku_image.tags = ['doku:latest']
    doku_container.image = doku_image

    # mock regular container
    regular_container = MagicMock(spec=Container)
    regular_container.short_id = 'cont1234'
    regular_container.name = 'container1'
    regular_container.attrs = {
        'Mounts': [{'Type': 'bind', 'Source': '/host/path', 'Destination': '/container/path', 'Mode': 'rw', 'RW': True}]
    }
    regular_image = MagicMock(spec=Image)
    regular_image.tags = ['image:latest']
    regular_container.image = regular_image

    # mock list containers
    mock_docker_client.containers.list.side_effect = lambda **kwargs: (
        [doku_container]
        if kwargs.get('filters') == {'label': 'github.repo=amerkurev/doku'}
        else [doku_container, regular_container, regular_container]
    )

    with (
        patch('scan.scanner.docker_from_env', return_value=mock_docker_client),
        patch('scan.scanner.doku_mounts', return_value=[docker_mount]),
        patch('scan.scanner.map_host_path_to_container') as mock_map_path,
        patch('scan.scanner.kvstore.set') as mock_kvstore_set,
    ):
        mock_path = MagicMock(spec=Path)
        mock_path.stat.return_value.st_size = 2048
        mock_map_path.return_value = mock_path

        scanner = BindMountsScanner(mock_is_stop)
        scanner.client = mock_docker_client
        scanner.scan()

        assert mock_docker_client.containers.list.call_count == 2
        mock_docker_client.containers.list.assert_any_call(filters={'label': 'github.repo=amerkurev/doku'})
        mock_docker_client.containers.list.assert_any_call(all=True)

        # check what was stored in the kvstore
        obj = DockerBindMounts(
            path='/host/path',
            err=False,
            size=0,
            scan_in_progress=False,
            last_scan='2023-01-01T12:00:00Z',
            containers=['container1', 'container1'],
        )
        obj.last_scan = ANY

        mock_kvstore_set.assert_has_calls(
            [
                call('/host/path', obj, ANY),
                call('/host/path', obj, ANY),
                call('/host/path', obj, ANY),
            ],
            any_order=True,
        )

        mock_map_path.return_value = None
        scanner.scan()


@pytest.fixture
def mock_diff_subdirs():
    counter = 0

    def _mock_func(path: Path) -> list[Path]:
        nonlocal counter
        counter += 1

        if counter == 1:  # len(subdirs) == 1
            m = MagicMock(spec=Path)
            m.name = 'root'
            return [m]
        elif counter == 3:  # len(subdirs) > 1
            m1 = MagicMock(spec=Path)
            m1.name = 'mnt'
            m2 = MagicMock(spec=Path)
            m2.name = 'root'
            return [m1, m2]
        return []  # len(subdirs) == 0

    return _mock_func


def test_overlay2_scanner(mock_docker_client, mock_is_stop, docker_mount, mock_diff_subdirs):
    # mock image data
    mock_image = MagicMock(spec=Image)
    mock_image.id = 'sha256:abc123'
    mock_image.tags = ['image:latest']
    mock_image.attrs = {
        'GraphDriver': {
            'Data': {
                'MergedDir': '/var/lib/docker/overlay2/img123/merged',
                'UpperDir': '/var/lib/docker/overlay2/img123/diff',
                'WorkDir': '/var/lib/docker/overlay2/img123/work',
                'LowerDir': '/var/lib/docker/overlay2/img456/diff:/var/lib/docker/overlay2/img789/diff',
            },
            'Name': 'overlay2',
        }
    }

    # mock container data
    mock_container = MagicMock(spec=Container)
    mock_container.short_id = 'cont1234'
    mock_container.name = 'container1'
    mock_container.attrs = {
        'GraphDriver': {
            'Data': {
                'MergedDir': '/var/lib/docker/overlay2/abc123/merged',
                'UpperDir': '/var/lib/docker/overlay2/abc123/diff',
                'WorkDir': '/var/lib/docker/overlay2/abc123/work',
                'LowerDir': '/var/lib/docker/overlay2/def456/diff:/var/lib/docker/overlay2/ghi789/diff',
            },
            'Name': 'overlay2',
        }
    }

    mock_container.image = mock_image

    mock_docker_client.containers.list.return_value = [mock_container]
    mock_docker_client.images.list.return_value = [mock_image]

    with (
        patch('scan.scanner.docker_from_env', return_value=mock_docker_client),
        patch('scan.scanner.doku_mounts', return_value=[docker_mount]),
        patch('scan.scanner.map_host_path_to_container') as mock_map_path,
        patch('pathlib.Path.exists', return_value=True),
        patch('scan.scanner.diff_subdirs', side_effect=mock_diff_subdirs),
        patch('scan.scanner.get_size') as mock_get_size,
        patch('scan.scanner.kvstore.set') as mock_kvstore_set,
    ):
        # mock for map_host_path_to_container
        mock_path = MagicMock(spec=Path)
        mock_path.stat.return_value.st_size = 1024
        mock_map_path.return_value = mock_path

        scanner = Overlay2Scanner(mock_is_stop)
        scanner.client = mock_docker_client

        scanner.overlay2_dir = MagicMock(spec=Path)

        overlay2_iterdir_0 = MagicMock(spec=Path)
        overlay2_iterdir_0.name = '/var/lib/docker/overlay2/img123'
        overlay2_iterdir_0.stat.return_value.st_ctime = 1234567890

        overlay2_iterdir_1 = MagicMock(spec=Path)
        overlay2_iterdir_1.name = '/var/lib/docker/overlay2/img456'
        overlay2_iterdir_1.stat.return_value.st_ctime = 1234567890

        scanner.overlay2_dir.iterdir.return_value = [overlay2_iterdir_0, overlay2_iterdir_1]

        mock_get_size.side_effect = [1024, Exception('Failed to get size')]
        scanner.scan()

        # verify method calls
        mock_docker_client.containers.list.assert_called_once_with(all=True)
        mock_docker_client.images.list.assert_called_once()

        overlay2_1 = DockerOverlay2Layer(
            id='/var/lib/docker/overlay2/img123',
            created='2023-01-01T12:00:00Z',
            diff_root='/root',
            err=False,
            size=1024,
            scan_in_progress=False,
            last_scan='2023-01-01T12:00:00Z',
            in_use=False,
        )
        overlay2_1.last_scan = ANY
        overlay2_1.created = ANY

        overlay2_2 = DockerOverlay2Layer(
            id='/var/lib/docker/overlay2/img456',
            created='2023-01-01T12:00:00Z',
            diff_root='/mnt, /root',
            err=True,
            size=0,
            scan_in_progress=False,
            last_scan='2023-01-01T12:00:00Z',
            in_use=False,
        )
        overlay2_2.last_scan = ANY
        overlay2_2.created = ANY

        # check what was stored in the kvstore
        mock_kvstore_set.assert_has_calls(
            [
                call('/var/lib/docker/overlay2/img123', overlay2_1, ANY),
                call('/var/lib/docker/overlay2/img123', overlay2_1, ANY),
                call('/var/lib/docker/overlay2/img456', overlay2_2, ANY),
                call('/var/lib/docker/overlay2/img456', overlay2_2, ANY),
            ],
            any_order=True,
        )

        # test when path mapping returns None
        mock_map_path.return_value = None
        scanner.scan()
