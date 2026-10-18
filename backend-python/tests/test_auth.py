import uuid

import pytest
from httpx import AsyncClient

from src.modules.auth.enums import AuthErrorCode

BASE_URL = "/api/v1/auth"
REGISTER = f"{BASE_URL}/register"
LOGIN = f"{BASE_URL}/login"
REFRESH = f"{BASE_URL}/refresh"
LOGOUT = f"{BASE_URL}/logout"
ME = f"{BASE_URL}/me"


@pytest.mark.asyncio
async def test_register_success(client: AsyncClient) -> None:
    unique_id = uuid.uuid4().hex[:8]
    payload = {
        "email": f"owner_{unique_id}@acme.com",
        "password": "SuperSecretPassword123!",
        "first_name": "Алексей",
        "last_name": "Смирнов",
        "organization_name": f"Acme {unique_id}",
    }

    response = await client.post(REGISTER, json=payload)

    assert response.status_code == 201
    body = response.json()
    assert body["ok"] is True
    assert body["error"] is None
    assert "access_token" in body["data"]
    assert body["data"]["token_type"] == "bearer"
    assert body["data"]["user"]["email"] == payload["email"]

    assert "refresh_token" in response.cookies


@pytest.mark.asyncio
async def test_register_duplicate_email_fails(client: AsyncClient) -> None:
    unique_id = uuid.uuid4().hex[:8]
    payload = {
        "email": f"duplicate_{unique_id}@acme.com",
        "password": "StrongPassword123!",
        "first_name": "Петр",
        "last_name": "Петров",
        "organization_name": "First Org",
    }
    first_response = await client.post(REGISTER, json=payload)
    assert first_response.status_code == 201

    second_response = await client.post(REGISTER, json=payload)
    assert second_response.status_code == 400
    body = second_response.json()
    assert body["ok"] is False
    assert body["data"] is None
    assert body["error"]["code"] == AuthErrorCode.EMAIL_ALREADY_EXISTS


@pytest.mark.asyncio
async def test_login_flow(client: AsyncClient) -> None:
    """Проверка входа: успешный сценарий и отказ при неверном пароле."""
    unique_id = uuid.uuid4().hex[:8]
    email = f"user_{unique_id}@acme.com"
    correct_password = "CorrectPassword123!"
    await client.post(
        REGISTER,
        json={
            "email": email,
            "password": correct_password,
            "first_name": "Сергей",
            "last_name": "Сидоров",
            "organization_name": "Login Test Org",
        },
    )

    wrong_login_response = await client.post(
        LOGIN,
        json={"email": email, "password": "WrongPassword9999!"},
    )
    assert wrong_login_response.status_code == 401
    assert wrong_login_response.json()["error"]["code"] == AuthErrorCode.INVALID_CREDENTIALS

    login_response = await client.post(
        LOGIN,
        json={"email": email, "password": correct_password},
    )
    assert login_response.status_code == 200
    body = login_response.json()
    assert body["ok"] is True
    assert "access_token" in body["data"]
    assert "refresh_token" in login_response.cookies


@pytest.mark.asyncio
async def test_get_me_protected_endpoint(client: AsyncClient) -> None:
    """Проверка авторизации: доступ к закрытому эндпоинту /me."""
    unique_id = uuid.uuid4().hex[:8]
    email = f"auth_me_{unique_id}@acme.com"
    reg_response = await client.post(
        "/api/v1/auth/register",
        json={
            "email": email,
            "password": "Password12345!",
            "first_name": "Анна",
            "last_name": "Каренина",
            "organization_name": "Anna Org",
        },
    )
    access_token = reg_response.json()["data"]["access_token"]
    # 1. Без токена доступа -> 401 Unauthorized
    unauth_response = await client.get("/api/v1/auth/me")
    assert unauth_response.status_code == 401
    # 2. С валидным Bearer-токеном -> 200 OK
    headers = {"Authorization": f"Bearer {access_token}"}
    auth_response = await client.get("/api/v1/auth/me", headers=headers)
    assert auth_response.status_code == 200
    assert auth_response.json()["data"]["email"] == email


@pytest.mark.asyncio
async def test_refresh_token_rotation(client: AsyncClient) -> None:
    """Проверка ротации сессий: старый refresh-токен заменяется новым."""
    unique_id = uuid.uuid4().hex[:8]
    reg_response = await client.post(
        "/api/v1/auth/register",
        json={
            "email": f"rotate_{unique_id}@acme.com",
            "password": "Password123!",
            "first_name": "Ольга",
            "last_name": "Романова",
            "organization_name": "Rotate Org",
        },
    )
    first_refresh_cookie = reg_response.cookies.get("refresh_token")
    assert first_refresh_cookie is not None
    # Вызываем /refresh, передавая cookie
    refresh_response = await client.post(
        "/api/v1/auth/refresh",
        cookies={"refresh_token": first_refresh_cookie},
    )
    assert refresh_response.status_code == 200
    second_refresh_cookie = refresh_response.cookies.get("refresh_token")
    # ИНВАРИАНТ РОТАЦИИ: новый токен сессии не равен старому!
    assert second_refresh_cookie is not None
    assert second_refresh_cookie != first_refresh_cookie


@pytest.mark.asyncio
async def test_replay_attack_detection(client: AsyncClient) -> None:
    """Критический тест безопасности: Детекция кражи токена (Replay Attack).

    Сценарий:
    1. Клиент регистрируется, получает Refresh Token R1.
    2. Клиент ротирует сессию: R1 аннулируется, выпускается R2.
    3. Злоумышленник пытается повторно отправить украденный R1.
    4. Сервер обязан:
       - Отклонить запрос со статусом 401 REPLAY_ATTACK_DETECTED.
       - Инвалидировать ВСЮ цепочку сессий пользователя (даже легитимный R2).
    """
    unique_id = uuid.uuid4().hex[:8]
    reg_response = await client.post(
        "/api/v1/auth/register",
        json={
            "email": f"hacker_{unique_id}@acme.com",
            "password": "Password123!",
            "first_name": "Виктор",
            "last_name": "Хакер",
            "organization_name": "Hacker Org",
        },
    )
    r1_token = reg_response.cookies.get("refresh_token")
    # Легитимный пользователь обновляет сессию
    refresh_response = await client.post(
        "/api/v1/auth/refresh",
        cookies={"refresh_token": r1_token},
    )
    assert refresh_response.status_code == 200
    r2_token = refresh_response.cookies.get("refresh_token")
    # АТАКА: Злоумышленник пытается войти по старому R1
    attack_response = await client.post(
        "/api/v1/auth/refresh",
        cookies={"refresh_token": r1_token},
    )
    assert attack_response.status_code == 401
    assert attack_response.json()["error"]["code"] == "REPLAY_ATTACK_DETECTED"
    # ПРОВЕРКА ЗАЩИТЫ: R2 тоже должен быть заблокирован сервером!
    blocked_response = await client.post(
        "/api/v1/auth/refresh",
        cookies={"refresh_token": r2_token},
    )
    assert blocked_response.status_code == 401
