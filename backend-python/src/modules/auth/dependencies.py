from fastapi import Depends
from fastapi.security import OAuth2PasswordBearer
from sqlalchemy.ext.asyncio import AsyncSession

from src.common.exceptions import UnauthorizedException
from src.core.database import get_db
from src.modules.auth.enums import AuthErrorCode
from src.modules.auth.models import User
from src.modules.auth.service import AuthService

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/v1/auth/login", auto_error=False)


def get_auth_service() -> AuthService:
    return AuthService()


async def get_current_user(
    token: str | None = Depends(oauth2_scheme),
    db: AsyncSession = Depends(get_db),
    auth_service: AuthService = Depends(get_auth_service),
) -> User:
    if not token:
        raise UnauthorizedException(
            message="Требуется аутентификация. Токен доступа не предоставлен",
            code=AuthErrorCode.INVALID_CREDENTIALS,
        )

    return await auth_service.get_current_user_by_token(db, token)
