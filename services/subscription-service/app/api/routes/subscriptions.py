import uuid

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from ..deps import get_db, get_current_user_id
from ..schemas.subscription import (
    CreateSubscriptionRequest,
    SubscriptionResponse,
)
from ...services.subscription_service import (
    SubscriptionService,
    SubscriptionNotFoundError,
    InvalidSubscriptionDataError,
    ActiveSubscriptionAlreadyExistsError,
    InvalidSubscriptionStatusError,
)

router = APIRouter(
    prefix="/api/v1/subscriptions",
    tags=["Subscriptions"],
)


@router.post(
    "/",
    response_model=SubscriptionResponse,
    status_code=status.HTTP_201_CREATED,
)
def create_subscription(
    payload: CreateSubscriptionRequest,
    user_id: uuid.UUID = Depends(get_current_user_id),
    db: Session = Depends(get_db),
):
    service = SubscriptionService(db)

    try:
        subscription = service.create_subscription(
            user_id=user_id,
            plan_id=payload.plan_id,
            frequency_days=payload.frequency_days,
        )
        return SubscriptionResponse.model_validate(subscription)
    except InvalidSubscriptionDataError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
    except ActiveSubscriptionAlreadyExistsError as e:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=str(e),
        )


@router.get(
    "/me",
    response_model=SubscriptionResponse,
)
def get_my_subscription(
    user_id: uuid.UUID = Depends(get_current_user_id),
    db: Session = Depends(get_db),
):
    service = SubscriptionService(db)

    try:
        subscription = service.get_my_active_subscription(user_id)
        return SubscriptionResponse.model_validate(subscription)
    except SubscriptionNotFoundError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )


@router.delete(
    "/{subscription_id}",
    response_model=SubscriptionResponse,
)
def cancel_subscription(
    subscription_id: uuid.UUID,
    user_id: uuid.UUID = Depends(get_current_user_id),
    db: Session = Depends(get_db),
):
    service = SubscriptionService(db)

    try:
        subscription = service.cancel_subscription(
            subscription_id=subscription_id,
            user_id=user_id,
        )
        return SubscriptionResponse.model_validate(subscription)
    except SubscriptionNotFoundError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )
    except InvalidSubscriptionStatusError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )