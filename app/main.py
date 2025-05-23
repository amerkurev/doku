import ssl
import logging
from pathlib import Path

from fastapi import FastAPI, status
from fastapi.requests import Request
from fastapi.responses import HTMLResponse, RedirectResponse
from fastapi.staticfiles import StaticFiles

import settings
from contrib.logger import setup_logger
from server.router import site
from server.state import lifespan


app = FastAPI(lifespan=lifespan, root_path=settings.ROOT_PATH)
app.mount('/static', StaticFiles(directory=settings.STATIC_DIR), name='static')
app.include_router(site.router)


@app.get('/', response_class=HTMLResponse, include_in_schema=False)
async def index(request: Request):
    url = request.url_for('dashboard')
    return RedirectResponse(url=url, status_code=status.HTTP_303_SEE_OTHER)


def main():
    logger = setup_logger()
    logger.info(f'Revision: {settings.REVISION}')

    if settings.LOG_LEVEL == logging.DEBUG:
        logger.debug(settings.to_string())

    kwargs = {
        'host': settings.HOST,
        'port': settings.PORT,
        'workers': settings.WORKERS,
        'reload': settings.DEBUG,
        'log_level': settings.LOG_LEVEL,
        'access_log': settings.DEBUG,
        'server_header': False,
        'ssl_cert_reqs': ssl.CERT_NONE,
        'ssl_ca_certs': None,  # TODO: add settings.SSL_CA_CERTS
        'ssl_ciphers': settings.SSL_CIPHERS,
        'proxy_headers': True,
        'forwarded_allow_ips': '*',
    }

    # enable SSL if key and cert files are provided
    if settings.SSL_KEYFILE and settings.SSL_CERTFILE:
        if Path(settings.SSL_KEYFILE).is_file() and Path(settings.SSL_CERTFILE).is_file():
            kwargs.update({
                'ssl_keyfile': settings.SSL_KEYFILE,
                'ssl_certfile': settings.SSL_CERTFILE,
            })
            if settings.SSL_KEYFILE_PASSWORD:
                kwargs['ssl_keyfile_password'] = settings.SSL_KEYFILE_PASSWORD

    import uvicorn

    uvicorn.run('main:app', **kwargs)


if __name__ == '__main__':
    main()  # pragma: no cover
