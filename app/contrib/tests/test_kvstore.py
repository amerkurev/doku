import datetime

import pytest
from peewee import SqliteDatabase
from playhouse.kv import KeyValue

from contrib import kvstore
from contrib.types import DockerImage


@pytest.fixture
def test_model():
    return DockerImage(
        Id='test_id',
        Created=datetime.datetime.now(),
        RepoTags=['test_repo:tag', 'test_repo:latest'],
        SharedSize=0,
        Size=100,
        containers=['test_container'],
    )


def test_set_get(test_model):
    db = SqliteDatabase(':memory:')
    kv = KeyValue(database=db, table_name='test_kvstore')

    with db:
        # Test setting and retrieving a single item
        key = 'test_key'
        kvstore.set(key, test_model, kv)

        ret = kvstore.get(key, kv, DockerImage)

        assert ret.id == test_model.id
        assert ret.created == test_model.created
        assert ret.repo_tags == test_model.repo_tags
        assert ret.shared_size == test_model.shared_size
        assert ret.size == test_model.size
        assert ret.containers == test_model.containers
        assert isinstance(ret, DockerImage)


def test_get_all(test_model):
    db = SqliteDatabase(':memory:')
    kv = KeyValue(database=db, table_name='test_kvstore')

    with db:
        # Test getting all items
        test_model.id = 'key1'
        kvstore.set('key1', test_model, kv)
        test_model.id = 'key2'
        kvstore.set('key2', test_model, kv)

        ret = kvstore.get_all(kv, DockerImage)

        assert len(ret) == 2
        for n, item in enumerate(ret, 1):
            assert item.id == 'key' + str(n)
            assert item.created == test_model.created
            assert item.repo_tags == test_model.repo_tags
            assert item.shared_size == test_model.shared_size
            assert item.size == test_model.size
            assert item.containers == test_model.containers
            assert isinstance(item, DockerImage)
