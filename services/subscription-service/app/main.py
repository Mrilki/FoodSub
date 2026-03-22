from fastapi import FastAPI

from .api.routes.subscriptions import router as subscriptions_router
from .api.routes.tarifs import router as tarifs_router
from .api.routes.admin_tarif import router as admin_tarif_router
from .api.routes.orders import router as orders_router

app = FastAPI(
    title="Subscription Service",
    version="1.0.0",
)

# ... после создания app, но до запуска uvicorn ...
def create_tables():
    Base.metadata.create_all(bind=engine)

app.include_router(subscriptions_router)
app.include_router(tarifs_router)
app.include_router(admin_tarif_router)
app.include_router(orders_router)


@app.get("/health")
def health_check():
    return {"status": "ok"}