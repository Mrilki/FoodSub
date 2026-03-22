import uuid

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from ..deps import get_db, require_admin
from ..schemas.plan import (
    CreatePlanRequest,
    PlanResponse,
)
from ...services.plan_service import (
    PlanService,
    PlanNotFoundError,
    PlanAlreadyExistsError,
    InvalidPlanDataError,
)

router = APIRouter(
    prefix="/api/v1/admin/tarif",
    tags=["Admin Tarifs"],
    dependencies=[Depends(require_admin)],
)


@router.post(
    "/",
    response_model=PlanResponse,
    status_code=status.HTTP_201_CREATED,
)
def create_tarif(
    payload: CreatePlanRequest,
    db: Session = Depends(get_db),
):
    service = PlanService(db)

    try:
        plan = service.create_plan(
            name=payload.name,
            price=payload.price,
        )
        return PlanResponse.model_validate(plan)
    except PlanAlreadyExistsError as e:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=str(e),
        )
    except InvalidPlanDataError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.delete(
    "/{plan_id}",
    status_code=status.HTTP_204_NO_CONTENT,
)
def delete_tarif(
    plan_id: uuid.UUID,
    db: Session = Depends(get_db),
):
    service = PlanService(db)

    try:
        service.delete_plan(plan_id)
    except PlanNotFoundError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )