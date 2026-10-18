from typing import Any

from fastapi import status


class AppException(Exception):
    def __init__(
        self,
        code: str,
        message: str,
        status_code: int = status.HTTP_400_BAD_REQUEST,
        details: dict[str, Any] | None = None,
    ) -> None:
        super().__init__(message)
        self.code = code
        self.message = message
        self.status_code = status_code
        self.details = details or {}


class BadRequestException(AppException):
    def __init__(
        self,
        message: str = "Bad request",
        code: str = "BAD_REQUEST",
    ) -> None:
        super().__init__(
            code=code,
            message=message,
            status_code=status.HTTP_400_BAD_REQUEST,
        )


class NotFoundException(AppException):
    def __init__(self, resource: str, identifier: Any, code: str = "NOT_FOUND") -> None:
        super().__init__(
            code=code,
            message=f"{resource} with id '{identifier}' not found",
            status_code=status.HTTP_404_NOT_FOUND,
            details={"resource": resource, "identifier": identifier},
        )


class UnauthorizedException(AppException):
    def __init__(
        self, message: str = "Authentication required", code: str = "UNAUTHORIZED"
    ) -> None:
        super().__init__(
            code=code,
            message=message,
            status_code=status.HTTP_401_UNAUTHORIZED,
        )


class ForbiddenException(AppException):
    def __init__(self, message: str = "Permission denied", code: str = "PERMISSION_DENIED") -> None:
        super().__init__(
            code=code,
            message=message,
            status_code=status.HTTP_403_FORBIDDEN,
        )


class ConflictException(AppException):
    def __init__(
        self, message: str, details: dict[str, Any] | None = None, code: str = "CONFLICT"
    ) -> None:
        super().__init__(
            code=code,
            message=message,
            status_code=status.HTTP_409_CONFLICT,
            details=details,
        )


class TenantAccessDeniedException(AppException):
    def __init__(
        self,
        message: str = "Access to the requested organization is forbidden",
        code: str = "TENANT_ACCESS_DENIED",
    ) -> None:
        super().__init__(
            code=code,
            message=message,
            status_code=status.HTTP_403_FORBIDDEN,
        )
