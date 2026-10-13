import secrets
import uuid
from datetime import UTC, datetime, timedelta

from jwt.exceptions import PyJWTError
from sqlalchemy.ext.asyncio import AsyncSession

from src.common.exceptions import BadRequestException, UnauthorizedException
from src.core.config import get_settings
from src.core.security import (
    create_access_token,
    decode_access_token,
    hash_password,
    hash_refresh_token,
    verify_password,
)
from src.modules.auth.enums import AuthErrorCode
from src.modules.auth.models import User
from src.modules.auth.repository import (
    OrganizationRepository,
    RefreshTokenRepository,
    UserRepository,
)
from src.modules.auth.schemas import (
    LoginRequest,
    RegisterRequest,
    TokenResponse,
    UserResponse,
)

settings = get_settings()


class AuthService:
    """Сервис аутентификации и авторизации."""

    def __init__(
        self,
        user_repo: UserRepository | None = None,
        org_repo: OrganizationRepository | None = None,
        refresh_repo: RefreshTokenRepository | None = None,
    ) -> None:
        self.user_repo = user_repo or UserRepository()
        self.org_repo = org_repo or OrganizationRepository()
        self.refresh_repo = refresh_repo or RefreshTokenRepository()

    def _generate_slug(self, name: str) -> str:
        """Генерирует базовый URL-совместимый слаг из названя компании.

        Например: 'Acme Corp' -> 'acme-corp'.
        Добавляет короткий случайный суффикс для гарантии уникальности.
        """
        base_slug = "".join(c if c.isalnum() else "-" for c in name.lower()).strip("-")
        base_slug = "-".join(filter(None, base_slug.split("-"))) or "org"
        random_suffix = secrets.token_hex(2)
        return f"{base_slug}-{random_suffix}"

    async def _issue_tokens(
        self,
        db: AsyncSession,
        user: User,
        user_agent: str | None = None,
        ip_address: str | None = None,
    ) -> tuple[TokenResponse, str]:
        """Выпуск новой пары токенов (Access + Refresh)."""
        access_token = create_access_token(
            subject=user.id,
            extra_claims={"email": user.email},
        )
        raw_refresh_token = secrets.token_hex(32)
        token_hash = hash_refresh_token(raw_refresh_token)
        expires_at = datetime.now(UTC) + timedelta(days=settings.REFRESH_TOKEN_EXPIRE_DAYS)
        await self.refresh_repo.create(
            db=db,
            user_id=user.id,
            token_hash=token_hash,
            expires_at=expires_at,
            user_agent=user_agent,
            ip_address=ip_address,
        )
        token_response = TokenResponse(
            access_token=access_token,
            token_type="bearer",
            user=UserResponse.model_validate(user),
        )

        return token_response, raw_refresh_token

    async def get_current_user_by_token(self, db: AsyncSession, token: str) -> User:
        try:
            payload = decode_access_token(token)
        except PyJWTError as err:
            raise UnauthorizedException(
                message="Недействительный или просроченный токен доступа",
                code=AuthErrorCode.INVALID_CREDENTIALS,
            ) from err

        user_id_raw: str | None = payload.get("sub")
        token_type: str | None = payload.get("type")

        if not user_id_raw or token_type != "access":
            raise UnauthorizedException(
                message="Неверный тип или формат токена",
                code=AuthErrorCode.INVALID_CREDENTIALS,
            )

        try:
            user_id = uuid.UUID(user_id_raw)
        except (ValueError, TypeError) as err:
            raise UnauthorizedException(
                message="Некорректный идентификатор пользователя в токене",
                code=AuthErrorCode.INVALID_CREDENTIALS,
            ) from err

        user = await self.user_repo.get_by_id(db, user_id)
        if not user:
            raise UnauthorizedException(
                message="Пользователь не найден.",
                code=AuthErrorCode.INVALID_CREDENTIALS,
            )

        if not user.is_active:
            raise UnauthorizedException(
                message="Аккаунт деактивирован администратором",
                code=AuthErrorCode.ACCOUNT_DISABLED,
            )

        return user

    async def register(
        self,
        db: AsyncSession,
        payload: RegisterRequest,
        user_agent: str | None = None,
        ip_address: str | None = None,
    ) -> tuple[TokenResponse, str]:
        """Сценарий регистрации нового пользователя и его компании.

        Возвращает:
        - TokenResponse (с access_token и user данными для JSON)
        - raw_refresh_token (строка токена для установки в HttpOnly cookie)
        """
        # 1. Проверяем инвариант
        existing_user = await self.user_repo.get_by_email(db, payload.email)
        if existing_user:
            raise BadRequestException(
                message="Пользователь с таким email уже зарегистрирован",
                code=AuthErrorCode.EMAIL_ALREADY_EXISTS,
            )

        # Создаем пользователя, организацию, назначем создателя владельцем (owner)
        user = await self.user_repo.create(
            db=db,
            email=payload.email,
            password_hash=hash_password(payload.password),
            first_name=payload.first_name,
            last_name=payload.last_name,
        )

        org = await self.org_repo.create(
            db=db,
            name=payload.organization_name,
            slug=self._generate_slug(payload.organization_name),
        )

        await self.org_repo.add_member(db=db, organization_id=org.id, user_id=user.id, role="owner")

        # Генерируем токены доступа и сессии
        tokens, raw_refresh = await self._issue_tokens(
            db=db, user=user, user_agent=user_agent, ip_address=ip_address
        )

        await db.commit()
        await db.refresh(user)
        return tokens, raw_refresh

    async def login(
        self,
        db: AsyncSession,
        payload: LoginRequest,
        user_agent: str | None,
        ip_address: str | None,
    ) -> tuple[TokenResponse, str]:
        """Сценарий входа по email и паролю."""
        user = await self.user_repo.get_by_email(db, payload.email)
        if not user:
            raise UnauthorizedException(
                message="Неверный email или пароль",
                code=AuthErrorCode.INVALID_CREDENTIALS,
            )

        if not user.is_active:
            raise UnauthorizedException(
                message="Аккаунт деактивирован администратором",
                code=AuthErrorCode.ACCOUNT_DISABLED,
            )

        if not verify_password(payload.password, user.password_hash):
            raise UnauthorizedException(
                message="Неверный email или пароль",
                code=AuthErrorCode.INVALID_CREDENTIALS,
            )

        tokens, raw_refresh = await self._issue_tokens(
            db=db,
            user=user,
            user_agent=user_agent,
            ip_address=ip_address,
        )

        await db.commit()

        return tokens, raw_refresh

    async def refresh_tokens(
        self,
        db: AsyncSession,
        raw_refresh_token: str,
        user_agent: str | None = None,
        ip_address: str | None = None,
    ) -> tuple[TokenResponse, str]:
        """Ротация токенов."""
        token_hash = hash_refresh_token(raw_refresh_token)
        token_record = await self.refresh_repo.get_by_hash(db, token_hash)

        if not token_record:
            raise UnauthorizedException(
                message="Недействительный refresh token",
                code=AuthErrorCode.INVALID_REFRESH_TOKEN,
            )

        if token_record.is_revoked:
            await self.refresh_repo.revoke_all_user_tokens(db=db, user_id=token_record.user_id)
            await db.commit()
            raise UnauthorizedException(
                message=(
                    "Обнаружена попытка повторного использования сессии. Выполните вход заново."
                ),
                code=AuthErrorCode.REPLAY_ATTACK_DETECTED,
            )

        if datetime.now(UTC) > token_record.expires_at:
            token_record.is_revoked = True
            await db.commit()
            raise UnauthorizedException(
                message="Срок действия сессии истек",
                code=AuthErrorCode.REFRESH_TOKEN_EXPIRED,
            )

        user = await self.user_repo.get_by_id(db, token_record.user_id)
        if not user or not user.is_active:
            raise UnauthorizedException(
                message="Пользователь не найден или заблокирован",
                code=AuthErrorCode.USER_INACTIVE,
            )

        await self.refresh_repo.revoke(db, token_record)

        tokens, new_raw_refresh = await self._issue_tokens(
            db=db,
            user=user,
            user_agent=user_agent,
            ip_address=ip_address,
        )

        await db.commit()
        return tokens, new_raw_refresh

    async def logout(self, db: AsyncSession, raw_refresh_token: str) -> None:
        """Сценарий выхода пользователя из системы. Отзываем сессию."""
        token_hash = hash_refresh_token(raw_refresh_token)
        token_record = await self.refresh_repo.get_by_hash(db, token_hash)
        if token_record:
            await self.refresh_repo.revoke(db, token_record)
            await db.commit()
