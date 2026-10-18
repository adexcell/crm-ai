from fastapi import APIRouter, Cookie, Depends, Request, Response, status
from sqlalchemy.ext.asyncio import AsyncSession

from src.common.envelope import ApiResponse
from src.common.exceptions import UnauthorizedException
from src.core.config import get_settings
from src.core.database import get_db
from src.modules.auth.dependencies import get_auth_service, get_current_user
from src.modules.auth.enums import AuthErrorCode
from src.modules.auth.models import User
from src.modules.auth.schemas import (
    LoginRequest,
    RegisterRequest,
    TokenResponse,
    UserResponse,
)
from src.modules.auth.service import AuthService

settings = get_settings()

router = APIRouter(prefix="/auth", tags=["auth"])


def _set_refresh_cookie(response: Response, raw_refresh_token: str) -> None:
    response.set_cookie(
        key="refresh_token",
        value=raw_refresh_token,
        httponly=True,
        secure=settings.is_production,
        samesite="lax",  # Защита от CSRF
        max_age=settings.REFRESH_TOKEN_EXPIRE_DAYS * 24 * 60 * 60,
        path="/",
    )


@router.post(
    "/register",
    response_model=ApiResponse[TokenResponse],
    status_code=status.HTTP_201_CREATED,
    summary="Регистрация нового пользователя и компании",
)
async def register(
    payload: RegisterRequest,
    request: Request,
    response: Response,
    db: AsyncSession = Depends(get_db),
    auth_service: AuthService = Depends(get_auth_service),
) -> ApiResponse[TokenResponse]:
    """Регистрирует пользователя, создает организацию, выпускает токены
    и устанавливает HttpOnly cookie с refresh-токеном.
    """
    user_agent = request.headers.get("user-agent")
    ip_address = request.client.host if request.client else None

    tokens, raw_refresh = await auth_service.register(
        db=db,
        payload=payload,
        user_agent=user_agent,
        ip_address=ip_address,
    )

    _set_refresh_cookie(response, raw_refresh)

    return ApiResponse(data=tokens)


@router.post(
    "/login",
    response_model=ApiResponse[TokenResponse],
    status_code=status.HTTP_200_OK,
    summary="Вход в систему по логину и паролю",
)
async def login(
    payload: LoginRequest,
    request: Request,
    response: Response,
    db: AsyncSession = Depends(get_db),
    auth_service: AuthService = Depends(get_auth_service),
) -> ApiResponse[TokenResponse]:
    user_agent = request.headers.get("user-agent")
    ip_address = request.client.host if request.client else None

    tokens, raw_refresh = await auth_service.login(
        db=db,
        payload=payload,
        user_agent=user_agent,
        ip_address=ip_address,
    )

    _set_refresh_cookie(response, raw_refresh)

    return ApiResponse(data=tokens)


@router.post(
    "/refresh",
    response_model=ApiResponse[TokenResponse],
    status_code=status.HTTP_200_OK,
    summary="Обновление пары токенов по HttpOnly cookie",
)
async def refresh_tokens(
    request: Request,
    response: Response,
    refresh_token: str | None = Cookie(default=None),
    db: AsyncSession = Depends(get_db),
    auth_service: AuthService = Depends(get_auth_service),
) -> ApiResponse[TokenResponse]:
    if not refresh_token:
        raise UnauthorizedException(
            message="Сессионный refresh token не найден в cookies",
            code=AuthErrorCode.INVALID_REFRESH_TOKEN,
        )

    user_agent = request.headers.get("user-agent")
    ip_address = request.client.host if request.client else None

    tokens, new_raw_refresh = await auth_service.refresh_tokens(
        db=db,
        raw_refresh_token=refresh_token,
        user_agent=user_agent,
        ip_address=ip_address,
    )

    _set_refresh_cookie(response, new_raw_refresh)

    return ApiResponse(data=tokens)


@router.post(
    "/logout",
    response_model=ApiResponse[dict[str, str]],
    status_code=status.HTTP_200_OK,
    summary="Выход: отзыв сессии в БД и удаление cookie",
)
async def logout(
    response: Response,
    refresh_token: str | None = Cookie(default=None),
    db: AsyncSession = Depends(get_db),
    auth_service: AuthService = Depends(get_auth_service),
) -> ApiResponse[dict[str, str]]:
    if refresh_token:
        await auth_service.logout(db=db, raw_refresh_token=refresh_token)

    response.delete_cookie(key="refresh_token", path="/")

    return ApiResponse(data={"message": "Успешный выход из системы"})


@router.get(
    "/me",
    response_model=ApiResponse[UserResponse],
    status_code=status.HTTP_200_OK,
    summary="Получение профиля текущего авторизованного пользователя",
)
async def get_me(
    current_user: User = Depends(get_current_user),
) -> ApiResponse[UserResponse]:
    return ApiResponse(data=UserResponse.model_validate(current_user))


# TODO:
# add routes
# post register/request
# post register/confirm
# patch me
# post forgot-password
# post reset-password
