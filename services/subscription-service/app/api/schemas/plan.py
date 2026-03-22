import uuid

from pydantic import BaseModel, Field, ConfigDict


class CreatePlanRequest(BaseModel):
    name: str = Field(..., min_length=2, max_length=100)
    price: float = Field(..., gt=0)


class PlanResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    name: str
    price: float


class AddPlanMenuItemRequest(BaseModel):
    menu_item_id: uuid.UUID


class PlanMenuItemResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    plan_id: uuid.UUID
    menu_item_id: uuid.UUID