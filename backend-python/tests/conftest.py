from collections.abc import AsyncGenerator

import pytest
from httpx import ASGITransport, AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.pool import NullPool

from src.core.config import get_settings
from src.core.database import get_db
from src.main import app

settings = get_settings()

# 1. Создаем тестовый движок с NullPool.
# NullPool не хранит открытые сокеты между тестами, что полностью
# исключает ошибку "attached to a different loop" в asyncpg.
test_engine = create_async_engine(
    settings.async_database_url,
    poolclass=NullPool,
)

test_session_factory = async_sessionmaker(
    bind=test_engine,
    class_=AsyncSession,
    expire_on_commit=False,
    autoflush=False,
)


# 2. Переопределяем зависимость get_db для тестовой среды
async def override_get_db() -> AsyncGenerator[AsyncSession, None]:
    async with test_session_factory() as session:
        try:
            yield session
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()


# Подменяем продовый генератор сессий на тестовый
app.dependency_overrides[get_db] = override_get_db


@pytest.fixture
async def client() -> AsyncGenerator[AsyncClient, None]:
    """Фикстура тестового HTTP-клиента.

    Каждый тест получает чистый изолированный клиент со свежими cookies.
    """
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://testserver") as ac:
        yield ac
