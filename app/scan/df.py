import time

import schedule

import settings
from scan.scanner import SystemDFScanner, LogfilesScanner
from contrib.signal import SignalHandler
from contrib.logger import setup_logger


def main():
    """
    DF scanner monitors disk space usage for Docker containers and log files.
    """
    signal_ = SignalHandler()
    logger = setup_logger()
    logger.info('DF scanner started (system df + logfiles).')

    # make sure the database file exists
    settings.DB_DF.parent.mkdir(parents=True, exist_ok=True)

    ### Docker Disk Usage Scanner ###
    scanner = SystemDFScanner()
    scanner.scan()  # run once immediately
    schedule.every(settings.SCAN_INTERVAL).seconds.do(scanner.scan)

    ### Logfiles Scanner ###
    scanner = LogfilesScanner(is_stop=signal_.is_stop)
    scanner.scan()  # run once immediately
    schedule.every(settings.SCAN_LOGFILE_INTERVAL).seconds.do(scanner.scan)

    # main loop
    while not signal_.is_stop():
        schedule.run_pending()
        time.sleep(1)

    logger.info('DF scanner stopped.')


if __name__ == '__main__':
    main()  # pragma: no cover
