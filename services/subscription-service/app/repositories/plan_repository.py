import uuid
from typing import Sequence

from sqlalchemy import select
from sqlalchemy.orm import Session, joinedload

from ..models.plan import Plan


class PlanRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(
        self,
        name: str,
        price: float,
    ) -> Plan:
        plan = Plan(
            name=name,
            price=price,
        )
        self.db.add(plan)
        self.db.flush()
        self.db.refresh(plan)
        return plan

    def get_by_id(self, plan_id: uuid.UUID) -> Plan | None:
        stmt = (
            select(Plan)
            .options(joinedload(Plan.menu_items))
            .where(Plan.id == plan_id)
        )
        return self.db.scalar(stmt)

    def get_by_name(self, name: str) -> Plan | None:
        stmt = select(Plan).where(Plan.name == name)
        return self.db.scalar(stmt)

    def list_all(self) -> Sequence[Plan]:
        stmt = select(Plan).order_by(Plan.price.asc(), Plan.name.asc())
        return self.db.scalars(stmt).all()

    def delete(self, plan: Plan) -> None:
        self.db.delete(plan)
        self.db.flush()