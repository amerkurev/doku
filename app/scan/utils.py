import shutil
import subprocess
import time
from collections.abc import Callable
from pathlib import Path
from subprocess import CompletedProcess

from humanize import naturalsize

import settings
from contrib.logger import get_logger


_files_processed = 0


def cpu_throttling(sleep_duration: float):
    global _files_processed
    _files_processed += 1
    if _files_processed % 100 == 0:  # every 100 files we sleep for a while
        time.sleep(sleep_duration)


def du_available() -> bool:
    """
    Check if the `du` command is available in the system.
    """
    return shutil.which('du') is not None


def run_du(path: Path) -> int:
    """
    Run the `du` command on a path and return the disk usage in bytes.
    """
    res: CompletedProcess = subprocess.run(['du', '-sb', path], capture_output=True, text=True)
    if res.returncode == 0:
        try:
            # output is in the format 'size path' (see -s option)
            return int(res.stdout.split()[0])
        except ValueError:
            pass

    # error branch
    logger = get_logger()
    output = repr(res.stderr or res.stdout).strip()
    logger.debug(f"Error running 'du' on {path}: {output}")
    return 0


def get_size(path: Path, /, sleep_duration: float, is_stop: Callable[[], bool], use_du=True) -> int:
    """
    Calculate disk usage of a path in bytes (recursively).
    Path can be a file or a directory.

    Args:
        path: Path to calculate size for
        sleep_duration: Duration to sleep for every 100 files processed
        is_stop: Callable to check if the process should stop
        use_du: Whether to use 'du' command
    """
    total = 0
    if is_stop():
        return total

    match path:
        case _ if path.is_dir(follow_symlinks=False):
            if use_du:
                total += run_du(path)
                cpu_throttling(sleep_duration)
            else:
                for item in path.iterdir():
                    total += get_size(
                        item,
                        sleep_duration=sleep_duration,
                        is_stop=is_stop,
                        use_du=use_du,
                    )

        case _ if path.is_file(follow_symlinks=False):
            total += path.stat(follow_symlinks=False).st_size
            cpu_throttling(sleep_duration)

    return total


def pretty_size(size: int) -> str:
    """
    Convert a size in bytes to a human-readable format.
    """
    if size == 0:
        return '0'
    binary = not settings.SI
    return naturalsize(size, binary=binary)
