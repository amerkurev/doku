from unittest.mock import patch

from settings import to_string as settings_to_string
from main import main


@patch('uvicorn.run')
def test_main_with_default_settings(mock_run):
    with patch('settings.SSL_KEYFILE', ''), patch('settings.SSL_CERTFILE', ''):
        main()
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        assert args[0] == 'main:app'
        assert 'ssl_keyfile' not in kwargs
        assert 'ssl_certfile' not in kwargs


@patch('uvicorn.run')
def test_main_with_ssl_files_not_exist(mock_run):
    with (
        patch('settings.SSL_KEYFILE', '/not/exist/key.pem'),
        patch('settings.SSL_CERTFILE', '/not/exist/cert.pem'),
        patch('pathlib.Path.is_file', return_value=False),
    ):
        main()
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        assert args[0] == 'main:app'
        assert 'ssl_keyfile' not in kwargs
        assert 'ssl_certfile' not in kwargs


@patch('uvicorn.run')
def test_main_with_ssl_files_exist(mock_run):
    with (
        patch('settings.SSL_KEYFILE', '/path/to/key.pem'),
        patch('settings.SSL_CERTFILE', '/path/to/cert.pem'),
        patch('settings.SSL_KEYFILE_PASSWORD', ''),
        patch('pathlib.Path.is_file', return_value=True),
    ):
        main()
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        assert args[0] == 'main:app'
        assert kwargs['ssl_keyfile'] == '/path/to/key.pem'
        assert kwargs['ssl_certfile'] == '/path/to/cert.pem'
        assert 'ssl_keyfile_password' not in kwargs


@patch('uvicorn.run')
def test_main_with_ssl_password(mock_run):
    with (
        patch('settings.SSL_KEYFILE', '/path/to/key.pem'),
        patch('settings.SSL_CERTFILE', '/path/to/cert.pem'),
        patch('settings.SSL_KEYFILE_PASSWORD', 'password'),
        patch('pathlib.Path.is_file', return_value=True),
    ):
        main()
        mock_run.assert_called_once()
        args, kwargs = mock_run.call_args
        assert args[0] == 'main:app'
        assert kwargs['ssl_keyfile'] == '/path/to/key.pem'
        assert kwargs['ssl_certfile'] == '/path/to/cert.pem'
        assert kwargs['ssl_keyfile_password'] == 'password'


def test_settings_to_string():
    with patch('settings._settings') as mock_settings:
        # Configure the mock settings object
        mock_settings.model_dump.return_value = {
            'host': '0.0.0.0',
            'port': 9090,
            'log_level': 'info',
            'ssl_keyfile': '/.ssl/key.pem',
            'ssl_keyfile_password': 'secret',
            'ssl_certfile': '/.ssl/cert.pem',
            'scan_interval': 60,
            'workers': 4,
            'docker_host': 'unix:///var/run/docker.sock',
            'git_tag': 'v1.2.3',
            'git_sha': 'abcdef1234567890',
        }

        # Set needed attributes on the mock
        mock_settings.ssl_keyfile_password = 'secret'

        # Call the function
        result = settings_to_string()

        # Verify the result contains expected information
        assert 'Doku settings:' in result

        # Verify each category exists
        assert 'General settings:' in result
        assert 'SSL settings:' in result
        assert 'Scan settings:' in result
        assert 'Uvicorn settings:' in result
        assert 'Docker settings:' in result
        assert 'Version info:' in result

        # Verify specific settings are included
        assert 'host: 0.0.0.0' in result
        assert 'port: 9090' in result
        assert 'log_level: info' in result

        # Verify password is masked
        assert 'ssl_keyfile_password: ********' in result
        assert 'ssl_keyfile_password: secret' not in result

        # Verify SSL settings
        assert 'ssl_keyfile: /.ssl/key.pem' in result
        assert 'ssl_certfile: /.ssl/cert.pem' in result

        # Verify Docker settings
        assert 'docker_host: unix:///var/run/docker.sock' in result

        # Verify version info
        assert 'git_tag: v1.2.3' in result
        assert 'git_sha: abcdef1234567890' in result
