import os
from pathlib import Path
from unittest.mock import patch, MagicMock

import settings
from scan.utils import du_available, cpu_throttling, run_du, get_size, pretty_size


def test_cpu_throttling():
    with patch('time.sleep') as mock_sleep:
        for _ in range(100):
            cpu_throttling(0.1)
        mock_sleep.assert_called_once_with(0.1)


def test_du_available():
    assert du_available() is True


def test_run_du():
    p = Path(__file__).resolve()
    s1 = run_du(Path(p))
    s2 = os.path.getsize(p)
    assert s1 == s2

    with patch('subprocess.run') as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout='not a number')
        assert run_du(Path(p)) == 0


def test_get_size():
    p = settings.BASE_DIR
    assert get_size(Path(p), 0.1, lambda: True, use_du=True) == 0
    s1 = get_size(Path(p), 0.1, lambda: False, use_du=True)
    s2 = get_size(Path(p), 0.1, lambda: False, use_du=False)
    assert round(s1, -6) == round(s2, -6)


def test_pretty_size():
    settings.SI = False
    assert pretty_size(0) == '0'
    assert pretty_size(1024) == '1.0 KiB'
    assert pretty_size(1048576) == '1.0 MiB'
    assert pretty_size(1073741824) == '1.0 GiB'

    settings.SI = True
    assert pretty_size(0) == '0'
    assert pretty_size(1000) == '1.0 kB'
    assert pretty_size(1000000) == '1.0 MB'
    assert pretty_size(1000000000) == '1.0 GB'
