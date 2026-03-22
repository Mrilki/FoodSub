# app/repositories/plan_menu_item_repository.py

import uuid
from typing import Sequence

from sqlalchemy import select
from sqlalchemy.orm import Session

from ..models.plan_menu_item import PlanMenuItem


class PlanMenuItemRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(
        self,
        plan_id: uuid.UUID,
        menu_item_id: uuid.UUID,
    ) -> PlanMenuItem:
        item = PlanMenuItem(
            plan_id=plan_id,
            menu_item_id=menu_item_id,
        )
        self.db.add(item)
        self.db.flush()
        self.db.refresh(item)
        return item

    def get_by_id(self, item_id: uuid.UUID) -> PlanMenuItem | None:
        stmt = select(PlanMenuItem).where(PlanMenuItem.id == item_id)
        return self.db.scalar(stmt)

    def get_by_plan_id(self, plan_id: uuid.UUID) -> Sequence[PlanMenuItem]:
        stmt = (
            select(PlanMenuItem)
            .where(PlanMenuItem.plan_id == plan_id)
            .order_by(PlanMenuItem.id.asc())
        )
        return self.db.scalars(stmt).all()

    def get_by_plan_and_menu_item(
        self,
        plan_id: uuid.UUID,
        menu_item_id: uuid.UUID,
    ) -> PlanMenuItem | None:
        stmt = (
            select(PlanMenuItem)
            .where(
                PlanMenuItem.plan_id == plan_id,
                PlanMenuItem.menu_item_id == menu_item_id,
            )
        )
        return self.db.scalar(stmt)

    def delete(self, item: PlanMenuItem) -> None:
        self.db.delete(item)
        self.db.flush()

    def delete_all_by_plan_id(self, plan_id: uuid.UUID) -> None:
        items = self.get_by_plan_id(plan_id)
        for item in items:
            self.db.delete(item)
        self.db.flush()