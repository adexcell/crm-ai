import uuid
from collections.abc import Sequence
from datetime import datetime

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from src.modules.auth.models import Organization, OrganizationMember, RefreshToken, User


class UserRepository:
    """Репозиторий для работы с таблицей пользователей (users)."""

    async def get_by_id(self, db: AsyncSession, user_id: uuid.UUID) -> User | None:
        stmt = (
            select(User)
            .where(User.id == user_id)
            .options(selectinload(User.memberships).selectinload(OrganizationMember.organization))
        )

        result = await db.execute(stmt)
        return result.scalar_one_or_none()

    async def get_by_email(self, db: AsyncSession, email: str) -> User | None:
        stmt = (
            select(User)
            .where(User.email == email)
            .options(selectinload(User.memberships).selectinload(OrganizationMember.organization))
        )

        result = await db.execute(stmt)
        return result.scalar_one_or_none()

    async def create(
        self,
        db: AsyncSession,
        email: str,
        password_hash: str,
        first_name: str,
        last_name: str,
    ) -> User:
        user = User(
            email=email.lower(),
            password_hash=password_hash,
            first_name=first_name.strip(),
            last_name=last_name.strip(),
        )
        db.add(user)
        await db.flush()
        return user


class OrganizationRepository:
    """Репозиторий для работы с организациями и членством (organizations)."""

    async def get_by_id(self, db: AsyncSession, id: uuid.UUID) -> Organization | None:
        stmt = select(Organization).where(Organization.id == id)
        result = await db.execute(stmt)
        return result.scalar_one_or_none()

    async def get_by_slug(self, db: AsyncSession, slug: str) -> Organization | None:
        stmt = select(Organization).where(Organization.slug == slug)
        result = await db.execute(stmt)
        return result.scalar_one_or_none()

    async def create(
        self,
        db: AsyncSession,
        name: str,
        slug: str,
    ) -> Organization:
        organization = Organization(
            name=name.strip(),
            slug=slug,
        )
        db.add(organization)
        await db.flush()
        return organization

    async def add_member(
        self,
        db: AsyncSession,
        organization_id: uuid.UUID,
        user_id: uuid.UUID,
        role: str = "member",
    ) -> OrganizationMember:
        member = OrganizationMember(
            organization_id=organization_id,
            user_id=user_id,
            role=role,
        )
        db.add(member)
        await db.flush()
        return member

    async def get_user_memberships(
        self, db: AsyncSession, user_id: uuid.UUID
    ) -> Sequence[OrganizationMember]:
        """Получение списка всех организаций, в которых состоит пользователь."""
        stmt = (
            select(OrganizationMember)
            .where(OrganizationMember.user_id == user_id)
            .options(selectinload(OrganizationMember.organization))
        )
        result = await db.execute(stmt)
        return result.scalars().all()


class RefreshTokenRepository:
    """Репозиторий для управления сессиями и refresh-токенами."""

    async def create(
        self,
        db: AsyncSession,
        user_id: uuid.UUID,
        token_hash: str,
        expires_at: datetime,
        user_agent: str | None = None,
        ip_address: str | None = None,
    ) -> RefreshToken:
        """Сохранение нового refresh-токена в базу данных."""
        refresh_token = RefreshToken(
            user_id=user_id,
            token_hash=token_hash,
            expires_at=expires_at,
            user_agent=user_agent,
            ip_address=ip_address,
        )
        db.add(refresh_token)
        await db.flush()
        return refresh_token

    async def get_by_hash(self, db: AsyncSession, token_hash: str) -> RefreshToken | None:
        """Поиск токена по его sha256 хешу."""
        stmt = select(RefreshToken).where(RefreshToken.token_hash == token_hash)
        result = await db.execute(stmt)
        return result.scalar_one_or_none()

    async def revoke(self, db: AsyncSession, token_record: RefreshToken) -> None:
        """Отзыв сессии (помечаем токен как отозванный)."""
        token_record.is_revoked = True
        await db.flush()

    async def revoke_all_user_tokens(self, db: AsyncSession, user_id: uuid.UUID) -> None:
        """Отзыв всех активных сессий пользователя (при подозрении на взлом)."""
        stmt = select(RefreshToken).where(
            RefreshToken.user_id == user_id,
            RefreshToken.is_revoked.is_(False),
        )
        result = await db.execute(stmt)
        tokens = result.scalars().all()
        for token in tokens:
            token.is_revoked = True
        await db.flush()
