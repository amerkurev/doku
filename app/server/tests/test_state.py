from unittest import mock

import pytest
from fastapi import FastAPI

from server import state
from server.state import lifespan


@pytest.fixture
def mock_settings():
    with mock.patch('server.state.settings') as mock_settings:
        mock_settings.VERSION = 'test_version'
        mock_settings.BASIC_HTPASSWD = None
        yield mock_settings


@pytest.mark.asyncio
async def test_lifespan_no_credentials(mock_settings):
    app = FastAPI()
    async with lifespan(app) as state_instance:
        assert state_instance['basic_auth_credentials'] is None
        assert state_instance['version'] == 'test_version'


@pytest.mark.asyncio
async def test_lifespan_with_credentials_file_not_found(mock_settings, tmp_path):
    mock_settings.BASIC_HTPASSWD = str(tmp_path / 'non_existent.htpasswd')

    app = FastAPI()
    async with lifespan(app) as state_instance:
        assert state_instance['basic_auth_credentials'] is None
        assert state_instance['version'] == 'test_version'


@pytest.mark.asyncio
async def test_lifespan_with_credentials_file(mock_settings, tmp_path):
    htpasswd_file = tmp_path / 'test.htpasswd'
    htpasswd_content = 'user1:hash1\nuser2:hash2'
    htpasswd_file.write_text(htpasswd_content)
    mock_settings.BASIC_HTPASSWD = str(htpasswd_file)

    app = FastAPI()
    async with lifespan(app) as state_instance:
        assert state_instance['basic_auth_credentials'] == {'user1': 'hash1', 'user2': 'hash2'}
        assert state_instance['version'] == 'test_version'


def test_state_typed_dict():
    # Test that State is a properly defined TypedDict
    state_dict = state.State(basic_auth_credentials={'user': 'hash'}, version='1.0.0')
    assert state_dict['basic_auth_credentials'] == {'user': 'hash'}
    assert state_dict['version'] == '1.0.0'
