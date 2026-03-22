from collections.abc import Generator
import uuid

from fastapi import Depends, Header, HTTPException, status
from sqlalchemy.orm import Session

from ..core.database import SessionLocal


def get_db() -> Generator[Session, None, None]:
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def get_current_user_id(
    authorization: str | None = Header(default=None),
) -> uuid.UUID:
    """
    Заглушка для JWT.
    Потом здесь будет:
    - проверка bearer token
    - декодирование JWT
    - извлечение user_id
    """
    if not authorization:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authorization header is required",
        )

    # Временная заглушка для локального теста
    # Заменится на реальную проверку JWT
    return uuid.UUID("11111111-1111-1111-1111-111111111111")


def require_admin(
    authorization: str | None = Header(default=None),
) -> None:
    """
    Заглушка проверки роли ADMIN.
    Потом здесь будет извлечение roles из JWT.
    """
    if not authorization:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authorization header is required",
        )

    # Временная заглушка
    # Здесь потом проверка роли ADMIN
    return None