from datetime import datetime, timezone

from contrib.types import (
    DockerVersion,
    DockerImage,
    DockerImageList,
    DockerContainer,
    DockerContainerList,
    DockerVolume,
    DockerVolumeList,
    DockerBuildCache,
    DockerBuildCacheList,
    DockerSystemDF,
    DockerMount,
    DockerContainerLog,
    DockerBindMounts,
    DockerOverlay2Layer,
    DiskUsage,
)


def test_docker_version():
    x = DockerVersion.model_validate({
        'Platform': {'Name': 'Docker Engine - Community'},
        'Version': '20.10.7',
        'ApiVersion': '1.41',
        'MinAPIVersion': '1.12',
        'Os': 'linux',
        'Arch': 'amd64',
        'KernelVersion': '5.10.25-linuxkit',
    })

    assert x.platform_name == 'Docker Engine - Community'
    assert x.version == '20.10.7'
    assert x.api_version == '1.41'
    assert x.min_api_version == '1.12'
    assert x.os == 'linux'
    assert x.arch == 'amd64'
    assert x.kernel_version == '5.10.25-linuxkit'


def test_docker_image():
    x = DockerImage.model_validate({
        'Id': 'sha256:123456789abcdef',
        'Created': '2023-01-01T12:00:00Z',
        'ParentId': 'sha256:987654321fedcba',
        'RepoTags': ['image:latest', 'image:v1'],
        'SharedSize': 1000,
        'Size': 5000,
    })

    assert x.id == 'sha256:123456789abcdef'
    assert x.created == datetime(2023, 1, 1, 12, 0, tzinfo=timezone.utc)
    assert x.parent_id == 'sha256:987654321fedcba'
    assert x.repo_tags == ['image:latest', 'image:v1']
    assert x.safe_repo_tags == x.repo_tags
    assert x.shared_size == 1000
    assert x.size == 5000
    assert x.short_id == x.id[7:19]
    assert isinstance(x.created_delta, str)
    assert x.containers == []

    # repo_tags = None
    x = DockerImage.model_validate({
        'Id': 'sha256:123456789abcdef',
        'Created': '2023-01-01T12:00:00Z',
        'ParentId': 'sha256:987654321fedcba',
        'RepoTags': None,
        'SharedSize': 1000,
        'Size': 5000,
    })

    assert x.repo_tags is None
    assert x.safe_repo_tags == ['<none>:<none>']

    # repo_tags = []
    x = DockerImage.model_validate({
        'Id': 'sha256:123456789abcdef',
        'Created': '2023-01-01T12:00:00Z',
        'ParentId': 'sha256:987654321fedcba',
        'RepoTags': [],
        'SharedSize': 1000,
        'Size': 5000,
    })

    assert x.repo_tags == []
    assert x.safe_repo_tags == ['<none>:<none>']


def test_docker_image_list():
    images = [
        {'Id': 'sha256:123', 'Created': '2023-01-01T12:00:00Z', 'Size': 1000},
        {'Id': 'sha256:456', 'Created': '2023-01-02T12:00:00Z', 'Size': 2000},
    ]
    x = DockerImageList.model_validate(images)

    assert len(x.root) == 2
    assert x[0].id == 'sha256:123'
    assert x[1].id == 'sha256:456'
    assert list(x) == x.root


def test_docker_container():
    x = DockerContainer.model_validate({
        'Id': '123456789abcdef',
        'Names': ['/container_name'],
        'Image': 'nginx:latest',
        'ImageID': 'sha256:abcdef123456',
        'Created': '2023-01-01T12:00:00Z',
        'SizeRw': 50000,
        'SizeRootFs': 200000,
        'State': 'running',
    })

    assert x.id == '123456789abcdef'
    assert x.names == ['/container_name']
    assert x.image == 'nginx:latest'
    assert x.image_id == 'sha256:abcdef123456'
    assert x.created == datetime(2023, 1, 1, 12, 0, tzinfo=timezone.utc)
    assert x.size_rw == 50000
    assert x.size_root_fs == 200000
    assert x.state == 'running'
    assert x.short_id == x.id[:12]
    assert x.clean_names == ['container_name']
    assert isinstance(x.created_delta, str)

    x = DockerContainer.model_validate({
        'Id': '123456789abcdef',
        'Image': 'nginx:latest',
        'ImageID': 'sha256:abcdef123456',
        'Created': '2023-01-01T12:00:00Z',
        'SizeRw': 50000,
        'SizeRootFs': 200000,
        'State': 'running',
    })
    assert x.clean_names == []


def test_docker_container_list():
    containers = [
        {
            'Id': '123',
            'Image': 'nginx:latest',
            'ImageID': 'sha256:abcdef123456abcde',
            'Created': '2023-01-01T12:00:00Z',
        },
        {
            'Id': '456',
            'Image': 'redis:latest',
            'ImageID': 'sha256:123456abcdef123456',
            'Created': '2023-01-02T12:00:00Z',
        },
    ]
    x = DockerContainerList.model_validate(containers)

    assert len(x.root) == 2
    assert x[0].id == '123'
    assert x[1].id == '456'
    assert list(x) == x.root


def test_docker_volume():
    x = DockerVolume.model_validate({
        'Name': 'volume1',
        'Driver': 'local',
        'CreatedAt': '2023-01-01T12:00:00Z',
        'Mountpoint': '/var/lib/docker/volumes/volume1',
        'Scope': 'local',
        'UsageData': {'Size': 1000, 'RefCount': 2},
    })

    assert x.name == 'volume1'
    assert x.driver == 'local'
    assert x.created_at == datetime(2023, 1, 1, 12, 0, tzinfo=timezone.utc)
    assert x.mountpoint == '/var/lib/docker/volumes/volume1'
    assert x.scope == 'local'
    assert x.size == 1000
    assert x.ref_count == 2
    assert x.containers == []


def test_docker_volume_list():
    volumes = [
        {'Name': 'vol1', 'Driver': 'local', 'CreatedAt': '2023-01-01T12:00:00Z', 'UsageData': {}},
        {'Name': 'vol2', 'Driver': 'local', 'CreatedAt': '2023-01-02T12:00:00Z', 'UsageData': {}},
    ]
    x = DockerVolumeList.model_validate(volumes)

    assert len(x.root) == 2
    assert x[0].name == 'vol1'
    assert x[0].ref_count == 0
    assert x[1].name == 'vol2'
    assert x[1].size == 0
    assert list(x) == x.root


def test_docker_build_cache():
    x = DockerBuildCache.model_validate({
        'ID': 'abcdef123456',
        'Type': 'layer',
        'Description': 'mount / from exec /bin/sh -c apk update && apk add ' * 10,
        'InUse': True,
        'Shared': False,
        'Size': 1024,
        'CreatedAt': '2023-01-01T12:00:00Z',
        'LastUsedAt': '2023-02-01T12:00:00Z',
        'UsageCount': 5,
    })

    assert x.id == 'abcdef123456'
    assert x.type == 'layer'
    assert x.description
    assert x.in_use is True
    assert x.shared is False
    assert x.size == 1024
    assert x.usage_count == 5
    assert isinstance(x.last_used_delta, str)
    assert x.short_desc == x.description[:180] + '...'


def test_docker_build_cache_list():
    caches = [
        {
            'ID': 'cache1',
            'Type': 'layer',
            'Description': 'Layer cache 1',
            'InUse': True,
            'Shared': False,
            'Size': 2048,
            'CreatedAt': '2023-01-01T12:00:00Z',
            'LastUsedAt': '2023-02-01T12:00:00Z',
            'UsageCount': 10,
        },
        {
            'ID': 'cache2',
            'Type': 'layer',
            'Description': 'Layer cache 2',
            'InUse': False,
            'Shared': True,
            'Size': 1024,
            'CreatedAt': '2023-01-02T12:00:00Z',
            'LastUsedAt': '2023-02-02T12:00:00Z',
            'UsageCount': 5,
        },
    ]

    x = DockerBuildCacheList.model_validate(caches)

    assert len(x.root) == 2
    assert x[0].id == 'cache1'
    assert x[1].id == 'cache2'
    assert x[0].in_use is True
    assert x[1].in_use is False
    assert x[0].size == 2048
    assert x[1].size == 1024
    assert x[0].usage_count == 10
    assert x[1].usage_count == 5
    assert isinstance(x[0].last_used_delta, str)
    assert isinstance(x[1].last_used_delta, str)
    assert list(x) == x.root


def test_docker_mount():
    x = DockerMount.model_validate({
        'Source': '/host/path',
        'Destination': '/container/path',
        'Mode': 'ro',
        'Propagation': 'rprivate',
        'RW': False,
        'Type': 'bind',
        'Root': True,
    })

    assert x.source == '/host/path'
    assert x.destination == '/container/path'
    assert x.mode == 'ro'
    assert x.propagation == 'rprivate'
    assert x.rw is False
    assert x.type == 'bind'
    assert x.root is True
    assert x.src == '/host/path'
    assert x.dst == '/container/path'
    assert x.info() == '/host/path -> /container/path:ro (root)'


def test_docker_container_log():
    x = DockerContainerLog.model_validate({
        'id': '123456789abc',
        'name': 'nginx',
        'image': 'nginx:latest',
        'path': '/var/lib/docker/containers/7d2de847bebae847b-json.log',
        'size': 50000,
        'last_scan': '2023-01-01T12:00:00Z',
    })

    assert x.id == '123456789abc'
    assert x.name == 'nginx'
    assert x.image == 'nginx:latest'
    assert x.path == '/var/lib/docker/containers/7d2de847bebae847b-json.log'
    assert x.size == 50000
    assert x.last_scan == datetime(2023, 1, 1, 12, 0, tzinfo=timezone.utc)
    assert isinstance(x.short_path, str)
    assert len(x.short_path) < len(x.path)


def test_docker_bind_mounts():
    x = DockerBindMounts.model_validate({
        'path': '/mnt/bind',
        'err': False,
        'size': 100000,
        'scan_in_progress': True,
        'last_scan': '2023-01-01T12:00:00Z',
        'containers': ['container1', 'container2'],
    })

    assert x.path == '/mnt/bind'
    assert x.err is False
    assert x.size == 100000
    assert x.scan_in_progress is True
    assert x.containers == ['container1', 'container2']
    assert isinstance(x.last_scan_delta, str)


def test_docker_overlay2_layer():
    x = DockerOverlay2Layer.model_validate({
        'id': 'jq1br13jcumomv9j6u8rce531485cee0f83624769a2d',
        'created': '2023-01-01T12:00:00Z',
        'diff_root': '/var/lib/docker/overlay2/diff/jq1br13jcumomv9j6u8rce531485cee0f83624769a2d',
        'err': False,
        'size': 5000,
        'scan_in_progress': False,
        'last_scan': '2023-02-01T12:00:00Z',
        'in_use': True,
    })

    assert x.id == 'jq1br13jcumomv9j6u8rce531485cee0f83624769a2d'
    assert x.diff_root == '/var/lib/docker/overlay2/diff/jq1br13jcumomv9j6u8rce531485cee0f83624769a2d'
    assert x.err is False
    assert x.size == 5000
    assert x.scan_in_progress is False
    assert x.in_use is True
    assert x.short_id == x.id[:22] + '...'
    assert isinstance(x.created_delta, str)
    assert isinstance(x.last_scan_delta, str)


def test_disk_usage():
    x = DiskUsage.model_validate({
        'total': 1000000,
        'used': 500000,
        'free': 500000,
        'percent': 50.0,
    })

    assert x.total == 1000000
    assert x.used == 500000
    assert x.free == 500000
    assert x.percent == 50.0
    assert isinstance(x.pretty_total, str)
    assert isinstance(x.pretty_used, str)
    assert isinstance(x.pretty_free, str)


def test_docker_system_df():
    data = {
        'Images': [
            {'Id': 'sha256:img1', 'Created': '2023-01-01T12:00:00Z', 'Size': 1500},
            {'Id': 'sha256:img2', 'Created': '2023-01-02T12:00:00Z', 'Size': 2500},
        ],
        'Containers': [
            {
                'Id': 'cont1',
                'Image': 'nginx:latest',
                'ImageID': 'sha256:abcdef123456abcde',
                'Created': '2023-01-01T12:00:00Z',
            },
            {
                'Id': 'cont2',
                'Image': 'redis:latest',
                'ImageID': 'sha256:123456abcdef123456',
                'Created': '2023-01-02T12:00:00Z',
            },
        ],
        'Volumes': [
            {'Name': 'vol1', 'Driver': 'local', 'CreatedAt': '2023-01-01T12:00:00Z'},
            {'Name': 'vol2', 'Driver': 'local', 'CreatedAt': '2023-01-02T12:00:00Z'},
        ],
        'BuildCache': [
            {
                'ID': 'cache1',
                'Type': 'layer',
                'Description': 'Layer cache 1',
                'InUse': True,
                'Shared': False,
                'Size': 2048,
                'CreatedAt': '2023-01-01T12:00:00Z',
                'LastUsedAt': '2023-02-01T12:00:00Z',
                'UsageCount': 10,
            }
        ],
    }

    x = DockerSystemDF.model_validate(data)

    assert len(x.images.root) == 2
    assert x.images[0].id == 'sha256:img1'
    assert x.images[1].id == 'sha256:img2'

    assert len(x.containers.root) == 2
    assert x.containers[0].id == 'cont1'
    assert x.containers[1].id == 'cont2'

    assert len(x.volumes.root) == 2
    assert x.volumes[0].name == 'vol1'
    assert x.volumes[1].name == 'vol2'

    assert len(x.build_cache.root) == 1
    assert x.build_cache[0].id == 'cache1'
    assert x.build_cache[0].size == 2048
    assert x.build_cache[0].usage_count == 10
