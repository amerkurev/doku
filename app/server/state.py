import contextlib
from pathlib import Path
from typing import TypedDict

from fastapi import FastAPI

import settings


class State(TypedDict):
    basic_auth_credentials: dict[str, str] | None  # username: bcrypt hash of password
    version: str


@contextlib.asynccontextmanager
async def lifespan(_: FastAPI):
    creds = None

    # load basic auth credentials from htpasswd file if provided
    if settings.BASIC_HTPASSWD:
        path = Path(settings.BASIC_HTPASSWD)
        if path.is_file():
            with path.open('r') as fd:
                creds = dict(line.strip().split(':', 1) for line in fd)

    yield State(
        basic_auth_credentials=creds,
        version=settings.VERSION,
    )
