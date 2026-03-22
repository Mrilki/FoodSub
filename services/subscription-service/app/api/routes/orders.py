import uuid

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from ..deps import get_db, get_current_user_id
from ..schemas.order import (
    OrderResponse,
    OrderDetailsResponse,
    OrderItemResponse,
)
from ...services.order_service import (
    OrderService,
    OrderNotFoundError,
    InvalidOrderStatusError,
)

router = APIRouter(
    prefix="/api/v1/orders",
    tags=["Orders"],
)


@router.get(
    "/",
    response_model=list[OrderResponse],
)
def get_my_orders(
    user_id: uuid.UUID = Depends(get_current_user_id),
    db: Session = Depends(get_db),
):
    service = OrderService(db)
    orders = service.get_user_orders(user_id)
    return [OrderResponse.model_validate(order) for order in orders]


@router.get(
    "/{order_id}",
    response_model=OrderDetailsResponse,
)
def get_order_details(
    order_id: uuid.UUID,
    user_id: uuid.UUID = Depends(get_current_user_id),
    db: Session = Depends(get_db),
):
    service = OrderService(db)

    try:
        order = service.get_user_order_details(
            user_id=user_id,
            order_id=order_id,
        )

        return OrderDetailsResponse(
            id=order.id,
            subscription_id=order.subscription_id,
            status=order.status,
            created_at=order.created_at,
            updated_at=order.updated_at,
            items=[OrderItemResponse.model_validate(item) for item in order.items],
        )
    except OrderNotFoundError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )


@router.post(
    "/{order_id}/cancel",
    response_model=OrderResponse,
)
def cancel_order(
    order_id: uuid.UUID,
    user_id: uuid.UUID = Depends(get_current_user_id),
    db: Session = Depends(get_db),
):
    service = OrderService(db)

    try:
        order = service.cancel_order(
            user_id=user_id,
            order_id=order_id,
        )
        return OrderResponse.model_validate(order)
    except OrderNotFoundError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )
    except InvalidOrderStatusError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )