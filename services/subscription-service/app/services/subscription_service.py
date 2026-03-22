import uuid
from datetime import date, timedelta

from sqlalchemy.orm import Session

from ..enums.subscription_status import SubscriptionStatus
from ..models.subscription import Subscription
from ..repositories.subscription_repository import SubscriptionRepository
from ..repositories.plan_repository import PlanRepository
from ..repositories.plan_menu_item_repository import PlanMenuItemRepository
from ..repositories.order_repository import OrderRepository
from ..repositories.order_item_repository import OrderItemRepository


class SubscriptionNotFoundError(Exception):
    pass


class InvalidSubscriptionDataError(Exception):
    pass


class ActiveSubscriptionAlreadyExistsError(Exception):
    pass


class InvalidSubscriptionStatusError(Exception):
    pass


class SubscriptionService:
    def __init__(self, db: Session):
        self.db = db
        self.subscription_repository = SubscriptionRepository(db)
        self.plan_repository = PlanRepository(db)
        self.plan_menu_item_repository = PlanMenuItemRepository(db)
        self.order_repository = OrderRepository(db)
        self.order_item_repository = OrderItemRepository(db)

    def create_subscription(
        self,
        user_id: uuid.UUID,
        plan_id: uuid.UUID,
        frequency_days: int = 2,
    ) -> Subscription:
        if frequency_days <= 0:
            raise InvalidSubscriptionDataError("frequency_days must be greater than 0")

        plan = self.plan_repository.get_by_id(plan_id)
        if not plan:
            raise InvalidSubscriptionDataError(f"Plan with id={plan_id} not found")

        active_subscription = self.subscription_repository.get_active_by_user_id(user_id)
        if active_subscription:
            raise ActiveSubscriptionAlreadyExistsError(
                f"User {user_id} already has an active subscription"
            )

        start_date = date.today()
        end_date = start_date + timedelta(days=7)

        subscription = self.subscription_repository.create(
            user_id=user_id,
            plan_id=plan_id,
            frequency_days=frequency_days,
            start_date=start_date,
            end_date=end_date,
            status=SubscriptionStatus.ACTIVE,
        )

        first_order = self.order_repository.create(
            subscription_id=subscription.id,
            delivery_date=start_date,
        )

        plan_menu_items = self.plan_menu_item_repository.get_by_plan_id(plan_id)
        menu_item_ids = [item.menu_item_id for item in plan_menu_items]

        if menu_item_ids:
            self.order_item_repository.bulk_create(
                order_id=first_order.id,
                menu_item_ids=menu_item_ids,
            )

        self.db.commit()
        self.db.refresh(subscription)
        return subscription

    def get_my_active_subscription(self, user_id: uuid.UUID) -> Subscription:
        subscription = self.subscription_repository.get_active_by_user_id(user_id)
        if not subscription:
            raise SubscriptionNotFoundError(
                f"Active subscription for user_id={user_id} not found"
            )
        return subscription

    def get_subscription_by_id(self, subscription_id: uuid.UUID) -> Subscription:
        subscription = self.subscription_repository.get_by_id(subscription_id)
        if not subscription:
            raise SubscriptionNotFoundError(
                f"Subscription with id={subscription_id} not found"
            )
        return subscription

    def cancel_subscription(
        self,
        subscription_id: uuid.UUID,
        user_id: uuid.UUID,
    ) -> Subscription:
        subscription = self.get_subscription_by_id(subscription_id)

        if subscription.user_id != user_id:
            raise SubscriptionNotFoundError(
                f"Subscription with id={subscription_id} not found for this user"
            )

        if subscription.status == SubscriptionStatus.CANCELLED:
            raise InvalidSubscriptionStatusError("Subscription already cancelled")

        updated = self.subscription_repository.update_status(
            subscription=subscription,
            status=SubscriptionStatus.CANCELLED,
        )
        self.db.commit()
        return updated

    def generate_orders_for_date(self, target_date: date) -> int:
        created_orders_count = 0
        subscriptions = self.subscription_repository.get_active_for_delivery_date(target_date)

        for subscription in subscriptions:
            days_from_start = (target_date - subscription.start_date).days
            if days_from_start < 0:
                continue

            if days_from_start % subscription.frequency_days != 0:
                continue

            already_exists = self.order_repository.exists_for_subscription_and_date(
                subscription_id=subscription.id,
                delivery_date=target_date,
            )
            if already_exists:
                continue

            order = self.order_repository.create(
                subscription_id=subscription.id,
                delivery_date=target_date,
            )

            plan_menu_items = self.plan_menu_item_repository.get_by_plan_id(subscription.plan_id)
            menu_item_ids = [item.menu_item_id for item in plan_menu_items]

            if menu_item_ids:
                self.order_item_repository.bulk_create(
                    order_id=order.id,
                    menu_item_ids=menu_item_ids,
                )

            created_orders_count += 1

        self.db.commit()
        return created_orders_count