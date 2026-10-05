import pytest
from httpx import AsyncClient


@pytest.mark.asyncio
async def test_liveness_endpoint(client: AsyncClient) -> None:
    response = await client.get("/api/v1/health/live")

    assert response.status_code == 200
    assert "x-request-id" in response.headers

    body = response.json()
    assert body["ok"] is True
    assert body["data"] == {"status": "alive"}
    assert body["error"] is None


@pytest.mark.asyncio
async def test_readiness_endpoint(client: AsyncClient) -> None:
    response = await client.get("/api/v1/health/ready")

    assert response.status_code == 200
    body = response.json()
    assert body["ok"] is True
    assert body["data"]["postgres"] == "ok"
    assert body["data"]["redis"] == "ok"
    assert body["error"] is None
