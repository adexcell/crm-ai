import logging
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.core.config import get_settings
from src.core.database import async_engine
from src.middleware.correlation import CorrelationIdMiddleware
from src.middleware.error_handler import setup_exception_handlers
from src.modules.health.routes import router as health_router

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] [%(name)s] %(message)s",
)
logger = logging.getLogger("crm.main")
settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    logger.info("Starting up %s in %s mode...", settings.APP_NAME, settings.APP_ENV)
    yield
    logger.info("Shutting down... Disposing database connections.")
    await async_engine.dispose()
    logger.info("Cleanup complete. Goodbye!")


def create_app() -> FastAPI:
    app = FastAPI(
        title=settings.APP_NAME,
        version="0.1.0",
        docs_url="/docs" if settings.DEBUG else None,
        redoc_url="/redoc" if settings.DEBUG else None,
        lifespan=lifespan,
    )

    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.CORS_ORIGINS,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    app.add_middleware(CorrelationIdMiddleware)

    setup_exception_handlers(app)

    app.include_router(health_router, prefix="/api/v1")

    return app

app = create_app()
