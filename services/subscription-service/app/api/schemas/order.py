import uuid
from datetime import date, datetime

from pydantic import BaseModel, ConfigDict

from ...enums.order_status import OrderStatus


class OrderItemResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    order_id: uuid.UUID
    menu_item_id: uuid.UUID


class OrderResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    subscription_id: uuid.UUID
    status: OrderStatus
    created_at: datetime
    updated_at: datetime


class OrderDetailsResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    subscription_id: uuid.UUID
    status: OrderStatus
    created_at: datetime
    updated_at: datetime
    items: list[OrderItemResponse]