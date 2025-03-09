from unittest.mock import patch, MagicMock

import pytest

from scan.du import main


@pytest.fixture
def mock_stop_signal():
    """Creates a mock that returns False 10 times, then True"""
    mock = MagicMock()
    mock.is_stop = MagicMock(side_effect=[False] * 10 + [True])
    return mock


@patch('scan.du.SignalHandler')
@patch('scan.du.setup_logger')
@patch('scan.du.BindMountsScanner')
@patch('scan.du.Overlay2Scanner')
@patch('scan.du.time.sleep')
@patch('scan.du.schedule.every')
def test_main(
    mock_schedule,
    mock_sleep,
    mock_overlay2_scanner,
    mock_bindmounts_scanner,
    mock_logger,
    mock_signal_handler,
    mock_stop_signal,
):
    # Setup mocks
    mock_signal_handler.return_value = mock_stop_signal
    mock_logger.return_value.info = MagicMock()

    # Run main function
    main()

    mock_signal_handler.assert_called_once()
    mock_logger.assert_called_once()

    mock_logger.return_value.info.assert_any_call('DU scanner started (bind mounts + overlay2).')
    mock_logger.return_value.info.assert_any_call('DU scanner stopped.')

    mock_overlay2_scanner.assert_called_once()
    mock_overlay2_scanner.return_value.scan.assert_called_once()

    mock_bindmounts_scanner.assert_called_once_with(is_stop=mock_stop_signal.is_stop)
    mock_bindmounts_scanner.return_value.scan.assert_called_once()

    assert mock_schedule.call_count == 2

    # Verify sleep was called 10 times (matches our mock signal setup)
    assert mock_sleep.call_count == 10
    mock_sleep.assert_called_with(1)
