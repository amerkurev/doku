import signal


class SignalHandler:
    """
    Signal handler for graceful shutdown of a scanner.
    """

    def __init__(self):
        self.running = True
        signal.signal(signal.SIGINT, self.handler)
        signal.signal(signal.SIGTERM, self.handler)

    def handler(self, sig, frame) -> None:
        self.running = False

    def is_stop(self) -> bool:
        return not self.running
