import logging

import settings
from uvicorn.logging import DefaultFormatter


LOGGER_NAME = 'doku'


def get_logger() -> logging.Logger:
    logger = logging.getLogger(LOGGER_NAME)
    return logger


def setup_logger() -> logging.Logger:
    """
    Configure a logger with settings from environment.
    """
    handler = logging.StreamHandler()

    # get the formatter from uvicorn for similar log output
    formatter = DefaultFormatter(
        fmt='%(levelprefix)s %(message)s',
        use_colors=not settings.IN_DOCKER,
    )
    handler.setFormatter(formatter)

    logger = logging.getLogger(LOGGER_NAME)
    logger.setLevel(settings.LOG_LEVEL)
    logger.handlers = []  # remove default handler
    logger.addHandler(handler)
    return logger
