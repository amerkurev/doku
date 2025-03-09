from unittest.mock import patch

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
