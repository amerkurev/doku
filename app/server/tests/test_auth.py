from unittest.mock import MagicMock

import bcrypt
import pytest
from fastapi import HTTPException, Request
from fastapi.security import HTTPBasicCredentials

from server.auth import check_password, basic_auth


def test_check_password_valid():
    password = 'testpassword'
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()
    assert check_password(password, hashed) is True


def test_check_password_invalid():
    password = 'testpassword'
    wrong_password = 'wrongpassword'
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()
    assert check_password(wrong_password, hashed) is False


def test_basic_auth_no_credentials():
    # Test when no credentials are configured in the app
    mock_request = MagicMock(spec=Request)
    mock_request.state.basic_auth_credentials = None
    credentials = HTTPBasicCredentials(username='someuser', password='somepass')

    # Should return username without checking password
    result = basic_auth(mock_request, credentials)
    assert result == 'someuser'


def test_basic_auth_valid_credentials():
    # Test with valid credentials
    password = 'testpassword'
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()

    mock_request = MagicMock(spec=Request)
    mock_request.state.basic_auth_credentials = {'testuser': hashed}
    credentials = HTTPBasicCredentials(username='testuser', password='testpassword')

    result = basic_auth(mock_request, credentials)
    assert result == 'testuser'


def test_basic_auth_invalid_username():
    # Test with invalid username
    password = 'testpassword'
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()

    mock_request = MagicMock(spec=Request)
    mock_request.state.basic_auth_credentials = {'testuser': hashed}
    credentials = HTTPBasicCredentials(username='wronguser', password='testpassword')

    with pytest.raises(HTTPException) as excinfo:
        basic_auth(mock_request, credentials)

    assert excinfo.value.status_code == 401
    assert excinfo.value.detail == 'Incorrect username or password'
    assert excinfo.value.headers == {'WWW-Authenticate': 'Basic'}


def test_basic_auth_invalid_password():
    # Test with invalid password
    password = 'testpassword'
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()

    mock_request = MagicMock(spec=Request)
    mock_request.state.basic_auth_credentials = {'testuser': hashed}
    credentials = HTTPBasicCredentials(username='testuser', password='wrongpassword')

    with pytest.raises(HTTPException) as excinfo:
        basic_auth(mock_request, credentials)

    assert excinfo.value.status_code == 401
    assert excinfo.value.detail == 'Incorrect username or password'
    assert excinfo.value.headers == {'WWW-Authenticate': 'Basic'}
