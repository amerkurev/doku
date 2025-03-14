import logging
from enum import Enum
from functools import cached_property
from pathlib import Path

from docker import constants as docker
from dotenv import load_dotenv
from pydantic import Field, PositiveInt, ValidationError, field_validator
from pydantic_settings import BaseSettings


load_dotenv()  # take environment variables from .env.


class LogLevel(str, Enum):
    DEBUG = 'debug'
    INFO = 'info'
    WARNING = 'warning'
    ERROR = 'error'
    CRITICAL = 'critical'


class ScanIntensity(str, Enum):
    AGGRESSIVE = 'aggressive'
    NORMAL = 'normal'
    LIGHT = 'light'


class Settings(BaseSettings):
    # general settings
    host: str = Field(alias='HOST', default='0.0.0.0', description='Interface address to bind the server to')
    port: PositiveInt = Field(alias='PORT', default=9090, description='Web interface port number')
    in_docker: bool = Field(alias='IN_DOCKER', default=False)
    log_level: LogLevel = Field(alias='LOG_LEVEL', default=LogLevel.INFO, description='Logging detail level')
    github_repo: str = Field(alias='GITHUB_REPO', default='amerkurev/doku')
    my_hostname: str = Field(alias='HOSTNAME', default='')  # it is set by the container automatically
    si: bool = Field(
        alias='SI', default=True, description='Use SI units (base 1000) instead of binary units (base 1024)'
    )
    basic_htpasswd: str = Field(
        alias='BASIC_HTPASSWD', default='/.htpasswd', description='Path to the htpasswd file for basic authentication'
    )

    # ssl settings
    ssl_keyfile: str = Field(alias='SSL_KEYFILE', default='/.ssl/key.pem')
    ssl_keyfile_password: str | None = Field(alias='SSL_KEYFILE_PASSWORD', default=None)
    ssl_certfile: str = Field(alias='SSL_CERTFILE', default='/.ssl/cert.pem')
    ssl_ciphers: str = Field(alias='SSL_CIPHERS', default='TLSv1')

    # scan settings
    scan_interval: PositiveInt = Field(
        alias='SCAN_INTERVAL', default=60, description='How often to collect basic Docker usage data (in seconds)'
    )
    scan_logfile_interval: PositiveInt = Field(
        alias='SCAN_LOGFILE_INTERVAL',
        default=60 * 5,
        description='How frequently to check container log sizes (in seconds)',
    )
    scan_bindmounts_interval: PositiveInt = Field(
        alias='SCAN_BINDMOUNTS_INTERVAL',
        default=60 * 60,
        description='Time between bind mount scanning operations (in seconds)',
    )
    scan_overlay2_interval: PositiveInt = Field(
        alias='SCAN_OVERLAY2_INTERVAL',
        default=60 * 60 * 24,
        description='How often to analyze Overlay2 storage (in seconds)',
    )
    scan_intensity: ScanIntensity = Field(
        alias='SCAN_INTENSITY',
        default=ScanIntensity.NORMAL,
        description='Performance impact level: "aggressive" (highest CPU usage), "normal" (balanced), or "light" (lowest impact)',
    )
    scan_use_du: bool = Field(
        alias='SCAN_USE_DU',
        default=True,
        description='Use the faster system `du` command for disk calculations instead of slower built-in methods',
    )

    # uvicorn settings
    workers: PositiveInt = Field(
        alias='UVICORN_WORKERS', default=1, description='Number of web server worker processes'
    )
    debug: bool = Field(alias='DEBUG', default=False, description='Enable debug mode')

    # docker daemon settings
    docker_host: str = Field(
        alias='DOCKER_HOST',
        default='unix:///var/run/docker.sock',
        description='Connection string for the Docker daemon',
    )
    docker_tls_verify: bool = Field(
        alias='DOCKER_TLS_VERIFY', default=False, description='Enable TLS verification for Docker daemon connection'
    )
    docker_cert_path: str | None = Field(
        alias='DOCKER_CERT_PATH', default=None, description='Directory containing Docker TLS certificates'
    )
    docker_version: str = Field(alias='DOCKER_VERSION', default='auto', description='Docker API version to use')
    docker_timeout: PositiveInt = Field(
        alias='DOCKER_TIMEOUT',
        default=docker.DEFAULT_TIMEOUT_SECONDS,
        description='Timeout in seconds for Docker API requests',
    )
    docker_max_pool_size: PositiveInt = Field(
        alias='DOCKER_MAX_POOL_SIZE',
        default=docker.DEFAULT_MAX_POOL_SIZE,
        description='Maximum number of connections in the Docker API connection pool',
    )
    docker_use_ssh_client: bool = Field(
        alias='DOCKER_USE_SSH_CLIENT',
        default=False,
        description='Use SSH for Docker daemon connection instead of HTTP/HTTPS',
    )

    # version settings
    git_tag: str = Field(alias='GIT_TAG', default='v0.0.0')
    git_sha: str = Field(alias='GIT_SHA', default='')

    @field_validator('log_level', mode='before')
    def lowercase_log_level(cls, v):
        if isinstance(v, str):
            return v.lower()
        return v

    @cached_property
    def log_level_num(self) -> int:
        level_map = {
            'debug': logging.DEBUG,
            'info': logging.INFO,
            'warning': logging.WARNING,
            'error': logging.ERROR,
            'critical': logging.CRITICAL,
        }
        return level_map[self.log_level]


try:
    _settings = Settings()
except ValidationError as err:
    raise SystemExit(err) from err


VERSION = _settings.git_tag
REVISION = f'{_settings.git_tag}-{_settings.git_sha[:7]}'

# general settings
HOST = _settings.host
PORT = _settings.port
IN_DOCKER = _settings.in_docker
LOG_LEVEL = _settings.log_level_num
GITHUB_REPO = _settings.github_repo
MY_HOSTNAME = _settings.my_hostname
SI = _settings.si
BASIC_HTPASSWD = _settings.basic_htpasswd
AUTH_ENABLED = Path(BASIC_HTPASSWD).exists()

# ssl settings
SSL_KEYFILE = _settings.ssl_keyfile
SSL_KEYFILE_PASSWORD = _settings.ssl_keyfile_password
SSL_CERTFILE = _settings.ssl_certfile
SSL_CIPHERS = _settings.ssl_ciphers

# scan settings
SCAN_INTERVAL = _settings.scan_interval
SCAN_LOGFILE_INTERVAL = _settings.scan_logfile_interval
SCAN_BINDMOUNTS_INTERVAL = _settings.scan_bindmounts_interval
SCAN_OVERLAY2_INTERVAL = _settings.scan_overlay2_interval
SCAN_INTENSITY = _settings.scan_intensity
SCAN_SLEEP_DURATION = {
    ScanIntensity.AGGRESSIVE: 0,  # no sleep, but CPU throttling
    ScanIntensity.NORMAL: 0.001,  # 1ms
    ScanIntensity.LIGHT: 0.01,  # 10ms
}[ScanIntensity(_settings.scan_intensity)]
SCAN_USE_DU = _settings.scan_use_du

# uvicorn settings
WORKERS = _settings.workers
DEBUG = _settings.debug

# docker daemon settings
DOCKER_HOST = _settings.docker_host
DOCKER_TLS_VERIFY = _settings.docker_tls_verify
DOCKER_CERT_PATH = _settings.docker_cert_path
DOCKER_VERSION = _settings.docker_version
DOCKER_TIMEOUT = _settings.docker_timeout
DOCKER_MAX_POOL_SIZE = _settings.docker_max_pool_size
DOCKER_USE_SSH_CLIENT = _settings.docker_use_ssh_client
DOCKER_ENV = {
    'DOCKER_HOST': DOCKER_HOST,
    'DOCKER_TLS_VERIFY': DOCKER_TLS_VERIFY or '',  # see kwargs_from_env in docker.from_env
    'DOCKER_CERT_PATH': DOCKER_CERT_PATH,
}

# paths
BASE_DIR = Path(__file__).resolve().parent
ROOT_DIR = BASE_DIR.parent
TEMPLATES_DIR = BASE_DIR / 'templates'
STATIC_DIR = BASE_DIR / 'static'
DB_DIR = BASE_DIR / 'db'
DB_DU = DB_DIR / 'du.sqlite3'
DB_DF = DB_DIR / 'df.sqlite3'
TABLE_LOGFILES = 'logfiles'
TABLE_BINDMOUNTS = 'bindmounts'
TABLE_SYSTEM_DF = 'system_df'
TABLE_OVERLAY2 = 'overlay2'
IMAGE_KEY = 'image'
CONTAINER_KEY = 'container'
VOLUME_KEY = 'volume'
BUILD_CACHE_KEY = 'build_cache'
ROOT_MOUNT_KEY = 'root_mount'


def to_string() -> str:
    """Return a formatted string with all settings for logging purposes."""
    settings_dict = _settings.model_dump()
    # Hide sensitive information
    if _settings.ssl_keyfile_password:
        settings_dict['ssl_keyfile_password'] = '********'

    lines = []
    lines.append('Doku settings:')

    # Group settings by categories for better readability
    categories = {
        'General settings': [
            'host',
            'port',
            'in_docker',
            'log_level',
            'github_repo',
            'my_hostname',
            'si',
            'basic_htpasswd',
        ],
        'SSL settings': ['ssl_keyfile', 'ssl_keyfile_password', 'ssl_certfile', 'ssl_ciphers'],
        'Scan settings': [
            'scan_interval',
            'scan_logfile_interval',
            'scan_bindmounts_interval',
            'scan_overlay2_interval',
            'scan_intensity',
            'scan_use_du',
        ],
        'Uvicorn settings': ['workers', 'debug'],
        'Docker settings': [
            'docker_host',
            'docker_tls_verify',
            'docker_cert_path',
            'docker_version',
            'docker_timeout',
            'docker_max_pool_size',
            'docker_use_ssh_client',
        ],
        'Version info': ['git_tag', 'git_sha'],
    }

    for category, keys in categories.items():
        lines.append(f'\n{category}:')
        for key in keys:
            if key in settings_dict:
                lines.append(f'  {key}: {settings_dict[key]}')

    return '\n'.join(lines)
