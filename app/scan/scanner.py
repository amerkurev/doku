import time
import fnmatch
from collections.abc import Callable
from pathlib import Path

from docker.models.containers import Container
from peewee import SqliteDatabase
from playhouse.kv import KeyValue
from pydantic import ValidationError

import settings
from scan.utils import get_size, du_available, pretty_size
from contrib import kvstore
from contrib.logger import get_logger
from contrib.types import (
    DockerSystemDF,
    DockerMount,
    DockerContainerLog,
    DockerBindMounts,
    DockerOverlay2Layer,
)
from contrib.docker import (
    docker_from_env,
    doku_container,
    doku_mounts,
    map_host_path_to_container,
)


class BaseScanner:
    def __init__(self):
        self.logger = get_logger()
        self.client = docker_from_env()

    def scan(self):
        raise NotImplementedError

    @property
    def database_name(self):
        raise NotImplementedError

    @property
    def table_name(self):
        raise NotImplementedError

    def log_start_time(self) -> Path:
        filename = settings.DB_DIR / f'{self.table_name}.timestamp'
        with filename.open('w') as fd:
            fd.write(str(int(time.time())))
        return filename


class SystemDFScanner(BaseScanner):
    """
    Scans the disk usage of the Docker system. E.g. images, containers, volumes.
    It's the equivalent of running `docker system df`.
    """

    @property
    def database_name(self):
        return settings.DB_DF

    @property
    def table_name(self):
        return settings.TABLE_SYSTEM_DF

    def scan(self):
        self.log_start_time()
        db = SqliteDatabase(self.database_name)
        kv = KeyValue(database=db, table_name=self.table_name)

        with db:
            start = time.perf_counter()
            self.logger.debug('Scanning Docker disk usage (df)...')

            data = self.client.df()
            df = DockerSystemDF.model_validate(data)

            # create image id -> image object mapping
            image_map = {item.id: item for item in df.images}

            # create volume name -> volume object mapping
            volume_map = {item.name: item for item in df.volumes}

            # add container name to each referenced image
            for cont in self.client.containers.list(all=True, sparse=False):
                image_id = cont.attrs['Image']
                mounts = cont.attrs.get('Mounts', [])

                # add container name to each referenced image
                if image_id in image_map:
                    image_map[image_id].containers.append(cont.name)

                # filter volume mounts and update volume objects
                volume_mounts = (
                    mount['Name'] for mount in mounts if mount['Type'] == 'volume' and mount['Name'] in volume_map
                )

                # add container name to each referenced volume
                for name in volume_mounts:
                    volume_map[name].containers.append(cont.name)

            kvstore.set(settings.IMAGE_KEY, df.images, kv)
            kvstore.set(settings.CONTAINER_KEY, df.containers, kv)
            kvstore.set(settings.VOLUME_KEY, df.volumes, kv)
            kvstore.set(settings.BUILD_CACHE_KEY, df.build_cache, kv)

            for mnt in doku_mounts(self.client):
                if mnt.root:
                    kvstore.set(settings.ROOT_MOUNT_KEY, mnt, kv)
                    break

            elapsed = time.perf_counter() - start
            self.logger.info(f'Docker disk usage (df) has been analyzed. Elapsed time: {elapsed:.2f} seconds.')


class LogfilesScanner(BaseScanner):
    """
    Scans the disk usage of log files.
    Docker stores log files in `/var/lib/docker/containers/<container-id>/`.
    """

    def __init__(self, is_stop: Callable[[], bool]):
        super().__init__()
        self.is_stop = is_stop
        self.root_mount = self._root_mount()

    def _root_mount(self) -> DockerMount | None:
        mounts = doku_mounts(self.client)
        root_mounts = [mnt for mnt in mounts if mnt.root]
        if not root_mounts:
            self.logger.error('No root mount found. Logfiles will not be scanned.')

        root_mount = root_mounts[0] if root_mounts else None
        return root_mount

    @property
    def database_name(self):
        return settings.DB_DF

    @property
    def table_name(self):
        return settings.TABLE_LOGFILES

    def scan(self):
        if not self.root_mount:
            return

        self.log_start_time()
        db = SqliteDatabase(self.database_name)
        kv = KeyValue(database=db, table_name=self.table_name)

        with db:
            total = 0
            num = 0
            start = time.perf_counter()
            self.logger.debug('Scanning logfiles...')

            kv.clear()  # clear previous calculations

            for cont in self.client.containers.list(all=True):
                if self.is_stop():
                    break

                cont: Container
                if 'LogPath' not in cont.attrs:
                    continue

                id_ = cont.short_id
                name = cont.name
                image_id = cont.image.short_id
                image = cont.image.tags[0] if cont.image.tags else image_id.removeprefix('sha256:')

                log_path = cont.attrs['LogPath']

                # map host path to doku container path (used only for size calculation)
                path: Path | None = map_host_path_to_container(
                    source=self.root_mount.src,
                    destination=self.root_mount.dst,
                    host_path=log_path,
                )

                if not path:
                    self.logger.error(f'Logfile {log_path} of container {name} not found or not accessible.')
                    continue

                # timestamp of the last scan in seconds
                last_scan = round(time.time())

                size = path.stat().st_size  # log file size in bytes
                total += size
                num += 1

                obj = DockerContainerLog(id=id_, name=name, image=image, path=log_path, size=size, last_scan=last_scan)
                kvstore.set(id_, obj, kv)
                self.logger.debug(f'Logfile of container {name} scanned. Size: {pretty_size(size)}.')

            elapsed = time.perf_counter() - start
            self.logger.info(
                f'{num} logfiles scanned. Total size: {pretty_size(total)}. Elapsed time: {elapsed:.2f} seconds.'
            )


class BindMountsScanner(BaseScanner):
    """
    Scans the disk usage of bind mounts.
    Bind mounts are used to share files between the host and a container.
    So we need to calculate the size of the files on the host.
    """

    def __init__(self, is_stop: Callable[[], bool]):
        super().__init__()
        self.is_stop = is_stop
        self.doku_mounts = self._doku_mounts()

    def _doku_mounts(self) -> list[DockerMount]:
        mounts = doku_mounts(self.client)

        if not mounts:
            self.logger.error('Doku container mounts not found. Bind mounts will not be scanned.')
        else:
            self.logger.info('Doku container mounts:')
            for mnt in mounts:
                self.logger.info(f'  {mnt.info()}')
        return mounts

    @property
    def database_name(self):
        return settings.DB_DU

    @property
    def table_name(self):
        return settings.TABLE_BINDMOUNTS

    def should_ignore_path(self, path: str) -> bool:
        """Check if the path matches any ignore pattern."""
        for pattern in settings.BINDMOUNT_IGNORE_PATTERNS:
            if fnmatch.fnmatch(path, pattern):
                return True
        return False

    def scan(self):
        if not self.doku_mounts:
            return

        self.log_start_time()
        db = SqliteDatabase(self.database_name)
        kv = KeyValue(database=db, table_name=self.table_name)

        with db:
            total = 0
            num = 0
            start = time.perf_counter()
            self.logger.info('Scanning bind mounts...')

            kv.clear()  # clear previous calculations

            already_scanned: dict[str, DockerBindMounts] = {}  # set of processed bindmounts
            myself = doku_container(self.client)

            # loop through all containers
            for cont in self.client.containers.list(all=True):
                cont: Container

                # skip containers without mounts
                if 'Mounts' not in cont.attrs:
                    continue

                # skip the current container
                if cont.id == myself.id:
                    continue

                name = cont.name
                mounts = cont.attrs['Mounts']

                # loop through all mounts of the container
                for mount in mounts:
                    if self.is_stop():
                        break

                    try:
                        mnt: DockerMount = DockerMount.model_validate(mount)
                    except ValidationError:
                        continue

                    # skip non-bind mounts
                    if mnt.type != 'bind':
                        continue

                    # skip mounts to docker socket and secrets
                    if mnt.dst == '/var/run/docker.sock' or mnt.dst.startswith('/run/secrets/'):
                        continue

                    # Skip paths matching ignore patterns
                    if self.should_ignore_path(mnt.src):
                        self.logger.debug(f'Skipping bind mount {mnt.src} as it matches an ignore pattern')
                        continue

                    if mnt.src in already_scanned:
                        # skip already scanned bind mounts, but update the list of containers
                        obj = already_scanned[mnt.src]
                        obj.containers.append(name)
                        kvstore.set(mnt.src, obj, kv)
                        continue

                    # timestamp of the last scan in seconds
                    last_scan = round(time.time())

                    obj = DockerBindMounts(
                        containers=[name],
                        path=mnt.src,
                        err=False,
                        size=0,
                        scan_in_progress=True,  # flag to indicate that the scan is in progress
                        last_scan=last_scan,
                    )
                    kvstore.set(mnt.src, obj, kv)  # for early access from the web interface
                    already_scanned[mnt.src] = obj

                    # looking for a suitable bind mount in doku_mounts
                    for doku_mnt in self.doku_mounts:
                        # map host path to doku container path (used only for size calculation)
                        path: Path | None = map_host_path_to_container(
                            source=doku_mnt.src,
                            destination=doku_mnt.dst,
                            host_path=mnt.src,
                        )

                        if not path:
                            continue

                        self.logger.debug(f'Start scanning bind mount {mnt.src} of container {name}...')
                        size = get_size(
                            path,
                            sleep_duration=settings.SCAN_SLEEP_DURATION,
                            is_stop=self.is_stop,
                            use_du=settings.SCAN_USE_DU and du_available(),
                        )
                        total += size
                        num += 1

                        obj.size = size
                        obj.scan_in_progress = False
                        kvstore.set(mnt.src, obj, kv)  # update the key-value store with the final size
                        already_scanned[mnt.src] = obj

                        self.logger.debug(f'Bind mount {mnt.src} scanned. Size: {pretty_size(size)}.')
                        # this mount is processed, no need to check other Doku mounts
                        break
                    else:
                        obj.err = True
                        obj.scan_in_progress = False
                        kvstore.set(mnt.src, obj, kv)  # update the key-value store with the error status
                        already_scanned[mnt.src] = obj
                        self.logger.error(f'Bind mount {mnt.src} of container {name} not found or not accessible.')

            elapsed = time.perf_counter() - start
            self.logger.info(
                f'{num} bind mounts scanned. Total size: {pretty_size(total)}. Elapsed time: {elapsed:.2f} seconds.'
            )


class Overlay2Scanner(BaseScanner):
    """
    Scans the disk usage of overlay2 storage driver.
    Docker stores overlay2 data in `/var/lib/docker/overlay2/`.
    """

    OVERLAY2_DIR = '/var/lib/docker/overlay2/'

    def __init__(self, is_stop: Callable[[], bool]):
        super().__init__()
        self.is_stop = is_stop
        self.overlay2_dir = self._overlay2_dir()

    def _overlay2_dir(self) -> Path | None:
        mounts = doku_mounts(self.client)
        root_mounts = [mnt for mnt in mounts if mnt.root]
        if not root_mounts:
            self.logger.error('No root mount found. Overlay2 storage driver will not be scanned.')

        root_mount = root_mounts[0] if root_mounts else None
        if not root_mount:
            return None

        return map_host_path_to_container(
            source=root_mount.src,
            destination=root_mount.dst,
            host_path=self.OVERLAY2_DIR,
        )

    @property
    def database_name(self):
        return settings.DB_DU

    @property
    def table_name(self):
        return settings.TABLE_OVERLAY2

    def collect_overlay2_layers(self) -> set[str]:
        layers = []
        graph = []

        # analyze graph driver of each container and image
        self.logger.debug('Collecting overlay2 layers from containers and images...')

        for cont in self.client.containers.list(all=True):
            if 'GraphDriver' in cont.attrs:
                graph.append(cont.attrs['GraphDriver'])

        for img in self.client.images.list(all=True):
            if 'GraphDriver' in img.attrs:
                graph.append(img.attrs['GraphDriver'])

        for g in graph:
            if 'Name' in g and g['Name'] == 'overlay2' and 'Data' in g:
                for key, item in g['Data'].items():
                    for path in item.split(':'):
                        if path.endswith('/diff') and path.startswith(self.OVERLAY2_DIR):
                            layers.append(Path(path).parent.name)

        layers = set(layers)
        self.logger.debug(f'Overlay2 layers: {len(layers)} collected.')
        return layers

    def scan(self):
        if not self.overlay2_dir:
            return

        self.log_start_time()
        db = SqliteDatabase(self.database_name)
        kv = KeyValue(database=db, table_name=self.table_name)
        layers = self.collect_overlay2_layers()

        with db:
            total = 0
            num = 0
            start = time.perf_counter()
            self.logger.info('Scanning overlay2 storage driver...')

            kv.clear()  # clear previous calculations

            for path in self.overlay2_dir.iterdir():
                if self.is_stop():
                    break

                diff_dir = path / 'diff'
                if not diff_dir.is_dir():
                    continue

                id_ = path.name
                short_id = id_[:12]
                created = path.stat().st_ctime

                # diff directories contain the actual data of the overlay2 layer.
                diff_root = Path('/')

                subdirs = diff_subdirs(diff_dir)
                if len(subdirs) == 0:
                    # empty diff directory, skip it
                    continue

                if len(subdirs) > 1:
                    # the root has many subdirectories, list them
                    diff_root = ', '.join('/' + x.name for x in subdirs)

                while len(subdirs) == 1:
                    # traverse the subdirectories while there is only one subdirectory on each level
                    diff_root /= subdirs[0].name
                    if subdirs[0].is_dir():
                        subdirs = diff_subdirs(subdirs[0])
                    else:
                        break  # the last element is a file

                # timestamp of the last scan in seconds
                last_scan = round(time.time())

                obj = DockerOverlay2Layer(
                    id=id_,
                    created=created,
                    diff_root=str(diff_root),
                    err=False,
                    size=0,
                    scan_in_progress=True,  # flag to indicate that the scan is in progress
                    last_scan=last_scan,
                    in_use=id_ in layers,
                )
                kvstore.set(id_, obj, kv)  # for early access from the web interface

                try:
                    self.logger.debug(f'Start scanning overlay2 layer {short_id}...')
                    # only diff directories are scanned
                    size = get_size(
                        diff_dir,
                        sleep_duration=settings.SCAN_SLEEP_DURATION,
                        is_stop=self.is_stop,
                        use_du=settings.SCAN_USE_DU and du_available(),
                    )
                    total += size
                    num += 1

                    obj.size = size
                    obj.scan_in_progress = False
                    kvstore.set(id_, obj, kv)  # update the key-value store with the final size
                    self.logger.debug(f'Overlay2 layer {short_id} scanned. Size: {pretty_size(size)}.')

                except Exception:
                    obj.err = True
                    obj.scan_in_progress = False
                    kvstore.set(id_, obj, kv)

            elapsed = time.perf_counter() - start
            self.logger.info(
                f'{num} overlay2 layers scanned. Total size: {pretty_size(total)}. Elapsed time: {elapsed:.2f} seconds.'
            )


def diff_subdirs(diff_dir: Path) -> list[Path]:
    return list(diff_dir.iterdir())
