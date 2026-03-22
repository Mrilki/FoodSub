from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..deps import get_db
from ..schemas.plan import PlanResponse
from ...services.plan_service import PlanService

router = APIRouter(
    prefix="/api/v1/tarifs",
    tags=["Tarifs"],
)


@router.get(
    "/",
    response_model=list[PlanResponse],
)
def list_tarifs(
    db: Session = Depends(get_db),
):
    service = PlanService(db)
    plans = service.list_plans()
    return [PlanResponse.model_validate(plan) for plan in plans]