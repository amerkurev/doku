import time

import schedule

import settings
from scan.scanner import BindMountsScanner, Overlay2Scanner
from contrib.signal import SignalHandler
from contrib.logger import setup_logger


def main():
    """
    DU scanner monitors disk space usage for Docker bind mounts and Docker overlay2 directory.
    """
    signal_ = SignalHandler()
    logger = setup_logger()
    logger.info('DU scanner started (bind mounts + overlay2).')

    # make sure the database file exists
    settings.DB_DU.parent.mkdir(parents=True, exist_ok=True)

    ### Bindmounts Scanner ###
    scanner = BindMountsScanner(is_stop=signal_.is_stop)
    scanner.scan()  # run once immediately
    schedule.every(settings.SCAN_BINDMOUNTS_INTERVAL).seconds.do(scanner.scan)

    ### Docker Overlay2 Scanner ###
    if settings.DISABLE_OVERLAY2_SCAN:
        logger.warning('Overlay2 scanner disabled.')
    else:
        scanner = Overlay2Scanner(is_stop=signal_.is_stop)
        scanner.scan()  # run once immediately
        schedule.every(settings.SCAN_OVERLAY2_INTERVAL).seconds.do(scanner.scan)

    # main loop
    while not signal_.is_stop():
        schedule.run_pending()
        time.sleep(1)

    logger.info('DU scanner stopped.')


if __name__ == '__main__':
    main()  # pragma: no cover
