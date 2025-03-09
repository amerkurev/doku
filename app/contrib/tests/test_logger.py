import logging

import pytest
from unittest.mock import patch, MagicMock

from contrib.logger import get_logger, setup_logger, LOGGER_NAME


def test_get_logger_returns_logger():
    logger = get_logger()
    assert isinstance(logger, logging.Logger)
    assert logger.name == LOGGER_NAME


@pytest.mark.parametrize('in_docker', [True, False])
@patch('contrib.logger.logging')
@patch('contrib.logger.settings')
@patch('contrib.logger.DefaultFormatter')
def test_setup_logger_configuration(
    mock_formatter_class,
    mock_settings,
    mock_logging,
    in_docker,
):
    mock_settings.IN_DOCKER = in_docker
    mock_settings.LOG_LEVEL = logging.INFO

    mock_handler = MagicMock()
    mock_logging.StreamHandler.return_value = mock_handler

    mock_formatter = MagicMock()
    mock_formatter_class.return_value = mock_formatter

    mock_logger = MagicMock()
    mock_logging.getLogger.return_value = mock_logger

    # Call the function
    result = setup_logger()

    # Assert the logger was configured correctly
    mock_logging.getLogger.assert_called_once_with(LOGGER_NAME)
    mock_logging.StreamHandler.assert_called_once()
    mock_formatter_class.assert_called_once_with(
        fmt='%(levelprefix)s %(message)s',
        use_colors=not in_docker,
    )
    mock_handler.setFormatter.assert_called_once_with(mock_formatter)
    mock_logger.setLevel.assert_called_once_with(mock_settings.LOG_LEVEL)
    assert mock_logger.handlers == []
    mock_logger.addHandler.assert_called_once_with(mock_handler)
    assert result == mock_logger
