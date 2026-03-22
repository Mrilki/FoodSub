import uuid

from sqlalchemy.orm import Session

from ..models.plan import Plan
from ..models.plan_menu_item import PlanMenuItem
from ..repositories.plan_repository import PlanRepository
from ..repositories.plan_menu_item_repository import PlanMenuItemRepository


class PlanNotFoundError(Exception):
    pass


class PlanAlreadyExistsError(Exception):
    pass


class PlanMenuItemAlreadyExistsError(Exception):
    pass


class PlanMenuItemNotFoundError(Exception):
    pass


class InvalidPlanDataError(Exception):
    pass


class PlanService:
    def __init__(self, db: Session):
        self.db = db
        self.plan_repository = PlanRepository(db)
        self.plan_menu_item_repository = PlanMenuItemRepository(db)

    def create_plan(self, name: str, price: float) -> Plan:
        if price <= 0:
            raise InvalidPlanDataError("Price must be greater than 0")

        existing = self.plan_repository.get_by_name(name)
        if existing:
            raise PlanAlreadyExistsError(f"Plan with name='{name}' already exists")

        plan = self.plan_repository.create(name=name, price=price)
        self.db.commit()
        return plan

    def get_plan_by_id(self, plan_id: uuid.UUID) -> Plan:
        plan = self.plan_repository.get_by_id(plan_id)
        if not plan:
            raise PlanNotFoundError(f"Plan with id={plan_id} not found")
        return plan

    def list_plans(self) -> list[Plan]:
        return list(self.plan_repository.list_all())

    def delete_plan(self, plan_id: uuid.UUID) -> None:
        plan = self.get_plan_by_id(plan_id)
        self.plan_repository.delete(plan)
        self.db.commit()

    def add_menu_item_to_plan(
        self,
        plan_id: uuid.UUID,
        menu_item_id: uuid.UUID,
    ) -> PlanMenuItem:
        self.get_plan_by_id(plan_id)

        existing = self.plan_menu_item_repository.get_by_plan_and_menu_item(
            plan_id=plan_id,
            menu_item_id=menu_item_id,
        )
        if existing:
            raise PlanMenuItemAlreadyExistsError(
                f"menu_item_id={menu_item_id} already exists in plan_id={plan_id}"
            )

        item = self.plan_menu_item_repository.create(
            plan_id=plan_id,
            menu_item_id=menu_item_id,
        )
        self.db.commit()
        return item

    def get_plan_menu_items(self, plan_id: uuid.UUID) -> list[PlanMenuItem]:
        self.get_plan_by_id(plan_id)
        return list(self.plan_menu_item_repository.get_by_plan_id(plan_id))

    def remove_menu_item_from_plan(
        self,
        plan_id: uuid.UUID,
        menu_item_id: uuid.UUID,
    ) -> None:
        self.get_plan_by_id(plan_id)

        item = self.plan_menu_item_repository.get_by_plan_and_menu_item(
            plan_id=plan_id,
            menu_item_id=menu_item_id,
        )
        if not item:
            raise PlanMenuItemNotFoundError(
                f"menu_item_id={menu_item_id} not found in plan_id={plan_id}"
            )

        self.plan_menu_item_repository.delete(item)
        self.db.commit()