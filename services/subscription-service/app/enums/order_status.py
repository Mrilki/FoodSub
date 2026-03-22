import enum


class OrderStatus(str, enum.Enum):
    SCHEDULED = "scheduled"
    ASSEMBLING = "assembling"
    DELIVERED = "delivered"
    CANCELLED = "cancelled"