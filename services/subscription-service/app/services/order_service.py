import uuid

from sqlalchemy.orm import Session

from ..enums.order_status import OrderStatus
from ..models.order import Order
from ..repositories.order_repository import OrderRepository


class OrderNotFoundError(Exception):
    pass


class InvalidOrderStatusError(Exception):
    pass


class OrderService:
    def __init__(self, db: Session):
        self.db = db
        self.order_repository = OrderRepository(db)

    def get_user_orders(self, user_id: uuid.UUID) -> list[Order]:
        return list(self.order_repository.get_user_orders(user_id))

    def get_user_order_details(
        self,
        user_id: uuid.UUID,
        order_id: uuid.UUID,
    ) -> Order:
        order = self.order_repository.get_user_order_by_id(
            user_id=user_id,
            order_id=order_id,
        )
        if not order:
            raise OrderNotFoundError(
                f"Order with id={order_id} not found for this user"
            )
        return order

    def cancel_order(
        self,
        user_id: uuid.UUID,
        order_id: uuid.UUID,
    ) -> Order:
        order = self.order_repository.get_user_order_by_id(
            user_id=user_id,
            order_id=order_id,
        )
        if not order:
            raise OrderNotFoundError(
                f"Order with id={order_id} not found for this user"
            )

        if order.status in {OrderStatus.ASSEMBLING, OrderStatus.DELIVERED, OrderStatus.CANCELLED}:
            raise InvalidOrderStatusError(
                f"Order in status={order.status} cannot be cancelled"
            )

        updated = self.order_repository.update_status(
            order=order,
            status=OrderStatus.CANCELLED,
        )
        self.db.commit()
        return updated