# app/repositories/subscription_repository.py

import uuid
from datetime import date
from typing import Sequence

from sqlalchemy import select
from sqlalchemy.orm import Session, joinedload

from ..models.subscription import Subscription
from ..enums.subscription_status import SubscriptionStatus


class SubscriptionRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(
        self,
        user_id: uuid.UUID,
        plan_id: uuid.UUID,
        start_date: date,
        end_date: date,
    ) -> Subscription:
        subscription = Subscription(
            user_id=user_id,
            plan_id=plan_id,
            start_date=start_date,
            end_date=end_date,
        )
        self.db.add(subscription)
        self.db.flush()
        self.db.refresh(subscription)
        return subscription

    def get_by_id(self, subscription_id: uuid.UUID) -> Subscription | None:
        stmt = (
            select(Subscription)
            .options(
                joinedload(Subscription.plan),
                joinedload(Subscription.orders),
            )
            .where(Subscription.id == subscription_id)
        )
        return self.db.scalar(stmt)

    def get_active_by_user_id(self, user_id: uuid.UUID) -> Subscription | None:
        stmt = (
            select(Subscription)
            .options(joinedload(Subscription.plan))
            .where(
                Subscription.user_id == user_id
            )
        )
        return self.db.scalar(stmt)

    def get_all_by_user_id(self, user_id: uuid.UUID) -> Sequence[Subscription]:
        stmt = (
            select(Subscription)
            .options(joinedload(Subscription.plan))
            .where(Subscription.user_id == user_id)
        )
        return self.db.scalars(stmt).all()

    def get_active_for_delivery_date(self, target_date: date) -> Sequence[Subscription]:
        stmt = (
            select(Subscription)
            .where(
                Subscription.start_date <= target_date,
                Subscription.end_date >= target_date,
            )
            .order_by(Subscription.start_date.asc())
        )
        return self.db.scalars(stmt).all()

    def delete(self, subscription: Subscription) -> None:
        self.db.delete(subscription)
        self.db.flush()