from collections.abc import Sequence
from datetime import datetime, UTC
from operator import attrgetter

import psutil
from humanize import naturaltime
from peewee import SqliteDatabase
from playhouse.kv import KeyValue
from pydantic import BaseModel

import settings
from contrib import kvstore
from contrib.docker import docker_from_env
from contrib.types import (
    DockerVersion,
    DockerImageList,
    DockerContainerList,
    DockerVolumeList,
    DockerBuildCacheList,
    DockerMount,
    DockerContainerLog,
    DockerBindMounts,
    DockerOverlay2Layer,
    DiskUsage,
)
from scan.utils import pretty_size


def total_size(items: Sequence | None, field_name='size') -> str:
    total = sum(map(attrgetter(field_name), items))
    return pretty_size(total)


def last_scan_time(table_name: str) -> tuple[datetime, str] | None:
    filename = settings.DB_DIR / f'{table_name}.timestamp'
    if filename.is_file():
        with filename.open('r') as fd:
            ts = int(fd.read().strip())
            dt = datetime.fromtimestamp(ts, UTC)
            return dt, naturaltime(dt)
    return None


def last_df_scan_time() -> tuple[datetime, str] | None:
    return last_scan_time(settings.TABLE_SYSTEM_DF)


def images() -> dict:
    items = None

    if settings.DB_DF.exists():
        db = SqliteDatabase(settings.DB_DF)
        with db:
            kv = KeyValue(database=db, table_name=settings.TABLE_SYSTEM_DF)
            if settings.IMAGE_KEY in kv:
                items = kvstore.get(settings.IMAGE_KEY, kv, DockerImageList).root

    context = {
        'name': 'images',
        'items': items,
        'total': total_size(items, field_name='shared_size'),
        'si': settings.SI,
        'last_scan_at': last_df_scan_time(),
    }
    return context


def containers() -> dict:
    items = None

    if settings.DB_DF.exists():
        db = SqliteDatabase(settings.DB_DF)
        with db:
            kv = KeyValue(database=db, table_name=settings.TABLE_SYSTEM_DF)
            if settings.CONTAINER_KEY in kv:
                items = kvstore.get(settings.CONTAINER_KEY, kv, DockerContainerList).root

    context = {
        'name': 'containers',
        'items': items,
        'total': total_size(items, field_name='size_rw'),
        'si': settings.SI,
        'last_scan_at': last_df_scan_time(),
    }
    return context


def volumes() -> dict:
    items = None

    if settings.DB_DF.exists():
        db = SqliteDatabase(settings.DB_DF)
        with db:
            kv = KeyValue(database=db, table_name=settings.TABLE_SYSTEM_DF)
            if settings.VOLUME_KEY in kv:
                items = kvstore.get(settings.VOLUME_KEY, kv, DockerVolumeList).root

    context = {
        'name': 'volumes',
        'items': items,
        'total': total_size(items),
        'si': settings.SI,
        'last_scan_at': last_df_scan_time(),
    }
    return context


def build_cache() -> dict:
    items = None

    if settings.DB_DF.exists():
        db = SqliteDatabase(settings.DB_DF)
        with db:
            kv = KeyValue(database=db, table_name=settings.TABLE_SYSTEM_DF)
            if settings.BUILD_CACHE_KEY in kv:
                items = kvstore.get(settings.BUILD_CACHE_KEY, kv, DockerBuildCacheList).root

    if items:
        for item in items:
            item.last_used = item.last_used.replace(microsecond=0)

    context = {
        'name': 'build cache',
        'items': items,
        'total': total_size(items),
        'si': settings.SI,
        'last_scan_at': last_df_scan_time(),
    }
    return context


def bind_mounts() -> dict:
    items = None

    if settings.DB_DU.exists():
        db = SqliteDatabase(settings.DB_DU)
        with db:
            kv = KeyValue(database=db, table_name=settings.TABLE_BINDMOUNTS)
            items = kvstore.get_all(kv, DockerBindMounts)

    context = {
        'name': 'bind mounts',
        'items': items,
        'total': total_size(items),
        'si': settings.SI,
        'last_scan_at': last_scan_time(settings.TABLE_BINDMOUNTS),
    }
    return context


def logs() -> dict:
    items = None

    if settings.DB_DF.exists():
        db = SqliteDatabase(settings.DB_DF)
        with db:
            kv = KeyValue(database=db, table_name=settings.TABLE_LOGFILES)
            items = kvstore.get_all(kv, DockerContainerLog)

    context = {
        'name': 'logs',
        'items': items,
        'total': total_size(items),
        'si': settings.SI,
        'last_scan_at': last_scan_time(settings.TABLE_LOGFILES),
    }
    return context


def overlay2() -> dict:
    items = None

    if settings.DB_DU.exists():
        db = SqliteDatabase(settings.DB_DU)
        with db:
            kv = KeyValue(database=db, table_name=settings.TABLE_OVERLAY2)
            items = kvstore.get_all(kv, DockerOverlay2Layer)

    context = {
        'name': 'overlay2',
        'items': items,
        'total': total_size(items),
        'si': settings.SI,
        'last_scan_at': last_scan_time(settings.TABLE_OVERLAY2),
    }
    return context


class Summary(BaseModel):
    num: int = 0
    total_size: int
    pretty_total_size: str


def summary(db: SqliteDatabase) -> dict[str, Summary]:
    r = {}

    # retrieve images, containers, volumes, and build cache
    kv = KeyValue(database=db, table_name=settings.TABLE_SYSTEM_DF)
    types = {
        settings.IMAGE_KEY: DockerImageList,
        settings.CONTAINER_KEY: DockerContainerList,
        settings.VOLUME_KEY: DockerVolumeList,
        settings.BUILD_CACHE_KEY: DockerBuildCacheList,
    }
    for key, value in types.items():
        if key in kv:
            items = kvstore.get(key, kv, value).root
            if key == settings.IMAGE_KEY:
                field_name = 'shared_size'
            elif key == settings.CONTAINER_KEY:
                field_name = 'size_rw'
            else:
                field_name = 'size'

            total_size = sum(map(attrgetter(field_name), items))
            r[key] = Summary(
                num=len(items),
                total_size=total_size,
                pretty_total_size=pretty_size(total_size),
            )

    # retrieve logs
    kv = KeyValue(database=db, table_name=settings.TABLE_LOGFILES)
    key = settings.TABLE_LOGFILES
    items = kvstore.get_all(kv, DockerContainerLog)
    total_size = sum(map(attrgetter('size'), items))
    r[key] = Summary(
        num=len(items),
        total_size=total_size,
        pretty_total_size=pretty_size(total_size),
    )
    return r


def overlay2_summary(db: SqliteDatabase) -> dict[str, Summary]:
    r = {}

    kv = KeyValue(database=db, table_name=settings.TABLE_OVERLAY2)
    layers = kvstore.get_all(kv, DockerOverlay2Layer)

    in_use = [layer for layer in layers if layer.in_use]
    not_in_use = [layer for layer in layers if not layer.in_use]

    in_use_size = sum(map(attrgetter('size'), in_use))
    not_in_use_size = sum(map(attrgetter('size'), not_in_use))

    r['overlay2_in_use'] = Summary(
        num=len(in_use),
        total_size=in_use_size,
        pretty_total_size=pretty_size(in_use_size),
    )
    r['overlay2_not_in_use'] = Summary(
        num=len(not_in_use),
        total_size=not_in_use_size,
        pretty_total_size=pretty_size(not_in_use_size),
    )
    r['overlay2'] = Summary(
        num=len(layers),
        total_size=in_use_size + not_in_use_size,
        pretty_total_size=pretty_size(in_use_size + not_in_use_size),
    )
    return r


def disk_usage() -> DiskUsage:
    du = psutil.disk_usage('/')

    if settings.DB_DF.exists():
        db = SqliteDatabase(settings.DB_DF)
        with db:
            # retrieve the root mount point
            kv = KeyValue(database=db, table_name=settings.TABLE_SYSTEM_DF)
            if settings.ROOT_MOUNT_KEY in kv:
                root_mount = kvstore.get(settings.ROOT_MOUNT_KEY, kv, DockerMount)
                _du = psutil.disk_usage(root_mount.destination)
                if _du.total > du.total:
                    du = _du

    return DiskUsage(
        total=du.total,
        used=du.used,
        free=du.free,
        percent=du.percent,
    )


def dashboard() -> dict:
    items = {}
    client = docker_from_env()
    version = DockerVersion.model_validate(client.version())

    if settings.DB_DF.exists():
        db = SqliteDatabase(settings.DB_DF)
        with db:
            # retrieve images, containers, volumes, build cache and logs
            items = summary(db)

    # pretty print the total size of each type
    total_size = sum(map(attrgetter('total_size'), items.values()))
    items['total'] = Summary(
        total_size=total_size,
        pretty_total_size=pretty_size(total_size),
    )

    if settings.DB_DU.exists():
        db = SqliteDatabase(settings.DB_DU)
        with db:
            # retrieve bind mounts and overlay2 layers
            items |= overlay2_summary(db)

    context = {
        'name': 'dashboard',
        'si': settings.SI,
        'version': version,
        'disk_usage': disk_usage(),
    } | items
    return context
