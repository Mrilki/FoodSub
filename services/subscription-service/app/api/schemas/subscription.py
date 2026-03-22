import uuid
from datetime import date, datetime

from pydantic import BaseModel, Field, ConfigDict

from ...enums.subscription_status import SubscriptionStatus


class CreateSubscriptionRequest(BaseModel):
    plan_id: uuid.UUID
    frequency_days: int = Field(default=2, ge=1)


class SubscriptionResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    user_id: uuid.UUID
    plan_id: uuid.UUID
    status: SubscriptionStatus
    frequency_days: int
    start_date: date
    end_date: date
    created_at: datetime
    updated_at: datetime


class SubscriptionWithPlanResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    user_id: uuid.UUID
    plan_id: uuid.UUID
    status: SubscriptionStatus
    frequency_days: int
    start_date: date
    end_date: date
    created_at: datetime
    updated_at: datetime
    plan_name: str
    plan_price: float