import redis.asyncio as aioredis
from fastapi import APIRouter, Depends, status
from fastapi.responses import JSONResponse
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession

from src.common.envelope import ApiResponse
from src.core.config import Settings, get_settings
from src.core.database import get_db

router = APIRouter(prefix="/health", tags=["Health"])


@router.get("/live", response_model=ApiResponse[dict[str, str]])
async def liveness() -> ApiResponse[dict[str, str]]:
    return ApiResponse.success({"status": "alive"})


@router.get("/ready")
async def readiness(
    db: AsyncSession = Depends(get_db),
    settings: Settings = Depends(get_settings),
) -> JSONResponse:
    checks: dict[str, str] = {}
    is_healthy = True

    try:
        await db.execute(text("SELECT 1"))
        checks["postgres"] = "ok"
    except Exception as exc:
        checks["postgres"] = f"error: {exc}"
        is_healthy = False

    try:
        client = aioredis.from_url(settings.redis_url, socket_timeout=2)
        await client.ping()
        await client.aclose()
        checks["redis"] = "ok"
    except Exception as exc:
        checks["redis"] = f"error: {exc}"
        is_healthy = False

    status_code = status.HTTP_200_OK if is_healthy else status.HTTP_503_SERVICE_UNAVAILABLE
    envelope = ApiResponse.success(checks) if is_healthy else ApiResponse.failure(
        code="SERVICE_UNAVAILABLE",
        message="One or more dependencies are unhealthy",
        details=checks,
    )
    return JSONResponse(status_code=status_code, content=envelope.model_dump(mode="json"))
