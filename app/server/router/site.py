from fastapi import APIRouter, Request
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates

import settings
from server.auth import AuthRequired, NoOpAuth
from server.router import context


if not settings.AUTH_ENABLED:
    AuthRequired = NoOpAuth


router = APIRouter(prefix='/site')
templates = Jinja2Templates(directory=settings.TEMPLATES_DIR)


@router.get('/', response_class=HTMLResponse, include_in_schema=False)
def dashboard(request: Request, _: AuthRequired):
    ctx = context.dashboard()
    return templates.TemplateResponse(request=request, name='pages/dashboard.html', context=ctx)


@router.get('/images/', response_class=HTMLResponse, include_in_schema=False)
def images(request: Request, _: AuthRequired):
    ctx = context.images()
    return templates.TemplateResponse(request=request, name='pages/images.html', context=ctx)


@router.get('/containers/', response_class=HTMLResponse, include_in_schema=False)
def containers(request: Request, _: AuthRequired):
    ctx = context.containers()
    return templates.TemplateResponse(request=request, name='pages/containers.html', context=ctx)


@router.get('/volumes/', response_class=HTMLResponse, include_in_schema=False)
def volumes(request: Request, _: AuthRequired):
    ctx = context.volumes()
    return templates.TemplateResponse(request=request, name='pages/volumes.html', context=ctx)


@router.get('/bind-mounts/', response_class=HTMLResponse, include_in_schema=False)
def bind_mounts(request: Request, _: AuthRequired):
    ctx = context.bind_mounts()
    return templates.TemplateResponse(request=request, name='pages/bind_mounts.html', context=ctx)


@router.get('/logs/', response_class=HTMLResponse, include_in_schema=False)
def logs(request: Request, _: AuthRequired):
    ctx = context.logs()
    return templates.TemplateResponse(request=request, name='pages/logs.html', context=ctx)


@router.get('/build-cache/', response_class=HTMLResponse, include_in_schema=False)
def build_cache(request: Request, _: AuthRequired):
    ctx = context.build_cache()
    return templates.TemplateResponse(request=request, name='pages/build_cache.html', context=ctx)


@router.get('/overlay2/', response_class=HTMLResponse, include_in_schema=False)
def overlay2(request: Request, _: AuthRequired):
    ctx = context.overlay2()
    return templates.TemplateResponse(request=request, name='pages/overlay2.html', context=ctx)
