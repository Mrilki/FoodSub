import uuid
from datetime import date
from typing import Sequence

from sqlalchemy import select
from sqlalchemy.orm import Session, joinedload

from ..models.order import Order
from ..models.subscription import Subscription
from ..enums.order_status import OrderStatus


class OrderRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(
        self,
        subscription_id: uuid.UUID,
        status: OrderStatus = OrderStatus.SCHEDULED,
    ) -> Order:
        order = Order(
            subscription_id=subscription_id,
            status=status,
        )
        self.db.add(order)
        self.db.flush()
        self.db.refresh(order)
        return order

    def get_by_id(self, order_id: uuid.UUID) -> Order | None:
        stmt = (
            select(Order)
            .options(
                joinedload(Order.items),
                joinedload(Order.subscription),
            )
            .where(Order.id == order_id)
        )
        return self.db.scalar(stmt)

    def get_by_subscription_id(self, subscription_id: uuid.UUID) -> Sequence[Order]:
        stmt = (
            select(Order)
            .where(Order.subscription_id == subscription_id)
        )
        return self.db.scalars(stmt).all()

    def get_user_orders(self, user_id: uuid.UUID) -> Sequence[Order]:
        stmt = (
            select(Order)
            .join(Subscription, Order.subscription_id == Subscription.id)
            .options(joinedload(Order.items))
            .where(Subscription.user_id == user_id)
        )
        return self.db.scalars(stmt).all()

    def get_user_order_by_id(self, user_id: uuid.UUID, order_id: uuid.UUID) -> Order | None:
        stmt = (
            select(Order)
            .join(Subscription, Order.subscription_id == Subscription.id)
            .options(
                joinedload(Order.items),
                joinedload(Order.subscription),
            )
            .where(
                Subscription.user_id == user_id,
                Order.id == order_id,
            )
        )
        return self.db.scalar(stmt)

    def exists_for_subscription(
        self,
        subscription_id: uuid.UUID,
    ) -> bool:
        stmt = select(Order).where(
            Order.subscription_id == subscription_id,
        )
        return self.db.scalar(stmt) is not None

    def update_status(
        self,
        order: Order,
        status: OrderStatus,
    ) -> Order:
        order.status = status
        self.db.add(order)
        self.db.flush()
        self.db.refresh(order)
        return order