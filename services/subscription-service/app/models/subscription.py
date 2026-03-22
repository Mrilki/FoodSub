import uuid
from datetime import datetime, date

from sqlalchemy import Date, DateTime, Integer, Enum, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from ..core.database import Base
from ..enums.subscription_status import SubscriptionStatus


class Subscription(Base):
    __tablename__ = "subscriptions"

    id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        primary_key=True,
        default=uuid.uuid4,
    )

    user_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        nullable=False,
        index=True,
    )

    plan_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("plans.id", ondelete="RESTRICT"),
        nullable=False,
        index=True,
    )

    # status: Mapped[SubscriptionStatus] = mapped_column(
    #     Enum(SubscriptionStatus),
    #     default=SubscriptionStatus.ACTIVE,
    #     nullable=False,
    #     index=True,
    # )

    start_date: Mapped[date] = mapped_column(
        Date,
        nullable=False,
    )

    end_date: Mapped[date] = mapped_column(
        Date,
        nullable=False,
    )

    plan = relationship("Plan", back_populates="subscriptions")
    orders = relationship("Order", back_populates="subscription", cascade="all, delete-orphan")