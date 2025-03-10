from datetime import datetime

from humanize import naturaltime
from pydantic import BaseModel, RootModel, Field

from scan.utils import pretty_size


def truncate(name: str, limit: int) -> str:
    if len(name) <= limit:
        return name
    return name[:limit] + '...'


class DockerVersion(BaseModel):
    platform_name: str = Field(alias=('Platform', 'Name'), default='')
    version: str = Field(alias='Version')
    api_version: str = Field(alias='ApiVersion')
    min_api_version: str = Field(alias='MinAPIVersion')
    os: str = Field(alias='Os')
    arch: str = Field(alias='Arch')
    kernel_version: str = Field(alias='KernelVersion')


class DockerImage(BaseModel):
    id: str = Field(alias='Id')
    created: datetime = Field(alias='Created')
    parent_id: str = Field(alias='ParentId', default='')
    repo_tags: list[str] | None = Field(alias='RepoTags', default_factory=list)
    shared_size: int = Field(alias='SharedSize', default=0)
    size: int = Field(alias='Size', default=0)
    containers: list[str] | None = Field(default_factory=list)

    @property
    def short_id(self) -> str:
        s = self.id.removeprefix('sha256:')
        return s[:12]

    @property
    def created_delta(self) -> str:
        return naturaltime(self.created)

    @property
    def safe_repo_tags(self) -> list[str]:
        if not self.repo_tags:  # empty list or None
            return ['<none>:<none>']
        return self.repo_tags


class DockerImageList(RootModel):
    root: list[DockerImage]

    def __iter__(self):
        return iter(self.root)

    def __getitem__(self, item):
        return self.root[item]


class DockerContainer(BaseModel):
    id: str = Field(alias='Id')
    names: list[str] | None = Field(alias='Names', default_factory=list)
    image: str = Field(alias='Image')
    image_id: str = Field(alias='ImageID')
    created: datetime = Field(alias='Created')
    size_rw: int = Field(alias='SizeRw', default=0)
    size_root_fs: int = Field(alias='SizeRootFs', default=0)
    state: str = Field(alias='State', default='')

    @property
    def short_id(self) -> str:
        return self.id[:12]

    @property
    def short_image(self) -> str:
        if self.image.startswith('sha256:'):
            return self.image[7:19]
        return self.image

    @property
    def clean_names(self) -> list[str]:
        if not self.names:
            return []
        return [name.lstrip('/') for name in self.names]

    @property
    def created_delta(self) -> str:
        return naturaltime(self.created)


class DockerContainerList(RootModel):
    root: list[DockerContainer]

    def __iter__(self):
        return iter(self.root)

    def __getitem__(self, item):
        return self.root[item]


class DockerVolume(BaseModel):
    name: str = Field(alias='Name')
    driver: str = Field(alias='Driver')
    created_at: datetime = Field(alias='CreatedAt')
    mountpoint: str = Field(alias='Mountpoint', default='')
    scope: str = Field(alias='Scope', default='local')
    usage_data: dict | None = Field(alias='UsageData', default_factory=dict)
    containers: list[str] | None = Field(default_factory=list)

    @property
    def short_name(self) -> str:
        return truncate(self.name, 39)

    @property
    def size(self) -> int:
        if 'Size' not in self.usage_data:
            return 0
        return self.usage_data['Size']

    @property
    def ref_count(self) -> int:
        if 'RefCount' not in self.usage_data:
            return 0
        return self.usage_data['RefCount']


class DockerVolumeList(RootModel):
    root: list[DockerVolume]

    def __iter__(self):
        return iter(self.root)

    def __getitem__(self, item):
        return self.root[item]


class DockerBuildCache(BaseModel):
    id: str = Field(alias='ID')
    type: str = Field(alias='Type')
    description: str = Field(alias='Description', default='')
    in_use: bool = Field(alias='InUse', default=False)
    shared: bool = Field(alias='Shared', default=False)
    size: int = Field(alias='Size', default=0)
    created_at: datetime = Field(alias='CreatedAt')
    last_used: datetime = Field(alias='LastUsedAt')
    usage_count: int = Field(alias='UsageCount', default=0)

    @property
    def last_used_delta(self) -> str:
        return naturaltime(self.last_used)

    @property
    def short_desc(self) -> str:
        return truncate(self.description, 180)


class DockerBuildCacheList(RootModel):
    root: list[DockerBuildCache]

    def __iter__(self):
        return iter(self.root)

    def __getitem__(self, item):
        return self.root[item]


class DockerSystemDF(BaseModel):
    images: DockerImageList = Field(alias='Images', default_factory=list)
    containers: DockerContainerList = Field(alias='Containers', default_factory=list)
    volumes: DockerVolumeList = Field(alias='Volumes', default_factory=list)
    build_cache: DockerBuildCacheList = Field(alias='BuildCache', default_factory=list)


class DockerMount(BaseModel):
    source: str = Field(alias='Source')
    destination: str = Field(alias='Destination')
    mode: str = Field(alias='Mode', default='')
    propagation: str = Field(alias='Propagation', default='')
    rw: bool = Field(alias='RW', default=False)
    type: str = Field(alias='Type', default='')
    root: bool = Field(alias='Root', default=False)

    @property
    def src(self) -> str:
        return self.source

    @property
    def dst(self) -> str:
        return self.destination

    def info(self) -> str:
        s = f'{self.src} -> {self.dst}'
        if self.mode:
            s += f':{self.mode}'
        if self.root:
            s += ' (root)'
        return s


class DockerContainerLog(BaseModel):
    id: str  # short ID of the container
    name: str  # name of the container
    image: str  # image of the container
    path: str  # path to the log file
    size: int  # size of the log file in bytes
    last_scan: datetime  # timestamp of the last scan

    @property
    def short_path(self) -> str:
        return self.path[:39] + '...' + self.path[-9:]


class DockerBindMounts(BaseModel):
    path: str  # path to the bind mount directory
    err: bool  # flag to indicate an error during scanning
    size: int  # size of the bind mount directory in bytes
    scan_in_progress: bool  # flag to indicate that the scan is in progress
    last_scan: datetime  # timestamp of the last scan
    containers: list[str]  # list of containers using the bind mount

    @property
    def last_scan_delta(self) -> str:
        return naturaltime(self.last_scan)


class DockerOverlay2Layer(BaseModel):
    id: str  # ID of the overlay2 layer
    created: datetime  # timestamp of the layer creation
    diff_root: str  # diff directory content
    err: bool  # flag to indicate an error during scanning
    size: int  # size of the overlay2 layer in bytes (only diff directory scanned)
    scan_in_progress: bool  # flag to indicate that the scan is in progress
    last_scan: datetime  # timestamp of the last scan
    in_use: bool  # flag to indicate that the layer is in use

    @property
    def short_id(self) -> str:
        return truncate(self.id, 22)

    @property
    def created_delta(self) -> str:
        return naturaltime(self.created)

    @property
    def last_scan_delta(self) -> str:
        return naturaltime(self.last_scan)


class DiskUsage(BaseModel):
    total: int
    used: int
    free: int
    percent: float

    @property
    def pretty_total(self) -> str:
        return pretty_size(self.total)

    @property
    def pretty_used(self) -> str:
        return pretty_size(self.used)

    @property
    def pretty_free(self) -> str:
        return pretty_size(self.free)
