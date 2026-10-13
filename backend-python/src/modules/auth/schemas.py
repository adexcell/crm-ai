import uuid
from datetime import datetime

from pydantic import BaseModel, ConfigDict, EmailStr, Field

from src.modules.auth.enums import OrganizationRole


class BaseDTO(BaseModel):
    model_config = ConfigDict(from_attributes=True)


class RegisterRequest(BaseModel):
    email: EmailStr
    password: str = Field(
        min_length=8, max_length=128, description="Пароль пользователя (минимум 8 символов)"
    )
    first_name: str = Field(min_length=1, max_length=100, description="Имя пользователя")
    last_name: str = Field(min_length=1, max_length=100, description="Фамилия пользователя")
    organization_name: str = Field(
        min_length=2,
        max_length=255,
        description="Название создаваемой компании (тенанта)",
    )


class LoginRequest(BaseModel):
    email: EmailStr
    password: str = Field(min_length=8, max_length=128)


class UserResponse(BaseDTO):
    id: uuid.UUID
    email: EmailStr
    first_name: str
    last_name: str
    is_active: bool
    is_superuser: bool
    created_at: datetime


class OrganizationResponse(BaseDTO):
    id: uuid.UUID
    name: str
    slug: str
    is_active: bool
    created_at: datetime


class OrganizationMemberResponse(BaseDTO):
    id: uuid.UUID
    organization_id: uuid.UUID
    user_id: uuid.UUID
    role: OrganizationRole
    joined_at: datetime
    organization: OrganizationResponse


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    user: UserResponse
