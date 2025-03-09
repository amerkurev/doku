import signal

from contrib.signal import SignalHandler


def test_signal_handler():
    handler = SignalHandler()
    handler.is_stop() is False
    handler.handler(signal.SIGINT, None)
    handler.is_stop() is True
