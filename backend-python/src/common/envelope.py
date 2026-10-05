from typing import Any, TypeVar

from pydantic import BaseModel, Field

T = TypeVar("T")


class ErrorDetail(BaseModel):
    code: str
    message: str
    details: dict[str, Any] = Field(default_factory=dict)


class ApiResponse[T](BaseModel):
    ok: bool
    data: T | None = None
    error: ErrorDetail | None = None

    @classmethod
    def success(cls, data: T) -> "ApiResponse[T]":
        return cls(ok=True, data=data, error=None)

    @classmethod
    def failure(
        cls, code: str, message: str, details: dict[str, Any] | None = None
    ) -> "ApiResponse[T]":
        return cls(
            ok=False,
            data=None,
            error=ErrorDetail(code=code, message=message, details=details or {}),
        )
