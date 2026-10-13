from enum import StrEnum


class OrganizationRole(StrEnum):
    """Роли пользователей внутри конкретной организации (RBAC)."""

    OWNER = "owner"  # Владелец компании (полный доступ, биллинг, удаление)
    ADMIN = "admin"  # Администратор (управление пользователями и настройками)
    MANAGER = "manager"  # Менеджер (работа со сделками, назначение ответственных)
    MEMBER = "member"  # Сотрудник (базовый доступ к назначенным задачам)


class AuthErrorCode(StrEnum):
    """Машиночитаемые коды ошибок для модуля аутентификации."""

    EMAIL_ALREADY_EXISTS = "EMAIL_ALREADY_EXISTS"
    INVALID_CREDENTIALS = "INVALID_CREDENTIALS"
    ACCOUNT_DISABLED = "ACCOUNT_DISABLED"
    INVALID_REFRESH_TOKEN = "INVALID_REFRESH_TOKEN"
    REPLAY_ATTACK_DETECTED = "REPLAY_ATTACK_DETECTED"
    REFRESH_TOKEN_EXPIRED = "REFRESH_TOKEN_EXPIRED"
    USER_INACTIVE = "USER_INACTIVE"
