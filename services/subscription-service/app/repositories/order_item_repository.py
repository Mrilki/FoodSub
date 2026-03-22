# app/repositories/order_item_repository.py

import uuid
from typing import Sequence

from sqlalchemy import select
from sqlalchemy.orm import Session

from ..models.order_item import OrderItem


class OrderItemRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(
        self,
        order_id: uuid.UUID,
        menu_item_id: uuid.UUID,
    ) -> OrderItem:
        item = OrderItem(
            order_id=order_id,
            menu_item_id=menu_item_id,
        )
        self.db.add(item)
        self.db.flush()
        self.db.refresh(item)
        return item

    def bulk_create(
        self,
        order_id: uuid.UUID,
        menu_item_ids: list[uuid.UUID],
    ) -> list[OrderItem]:
        items: list[OrderItem] = []
        for menu_item_id in menu_item_ids:
            item = OrderItem(
                order_id=order_id,
                menu_item_id=menu_item_id,
            )
            self.db.add(item)
            items.append(item)

        self.db.flush()

        for item in items:
            self.db.refresh(item)

        return items

    def get_by_order_id(self, order_id: uuid.UUID) -> Sequence[OrderItem]:
        stmt = (
            select(OrderItem)
            .where(OrderItem.order_id == order_id)
            .order_by(OrderItem.id.asc())
        )
        return self.db.scalars(stmt).all()

    def delete_all_by_order_id(self, order_id: uuid.UUID) -> None:
        items = self.get_by_order_id(order_id)
        for item in items:
            self.db.delete(item)
        self.db.flush()