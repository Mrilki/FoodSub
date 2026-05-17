import json
import logging
from typing import Any

from aiokafka import AIOKafkaProducer

logger = logging.getLogger(__name__)


class KafkaProducerService:
    def __init__(self, bootstrap_servers: str):
        self.bootstrap_servers = bootstrap_servers
        self._producer: AIOKafkaProducer | None = None

    async def start(self) -> None:
        if self._producer is not None:
            return

        self._producer = AIOKafkaProducer(
            bootstrap_servers=self.bootstrap_servers,
            key_serializer=self._serialize_key,
            value_serializer=self._serialize_value,
        )

        await self._producer.start()
        logger.info("Kafka producer started")

    async def stop(self) -> None:
        if self._producer is None:
            return

        await self._producer.stop()
        self._producer = None
        logger.info("Kafka producer stopped")

    async def send(
        self,
        topic: str,
        value: dict[str, Any],
        key: str | None = None,
    ) -> None:
        if self._producer is None:
            raise RuntimeError("Kafka producer is not started")

        await self._producer.send_and_wait(
            topic=topic,
            key=key,
            value=value,
        )

        logger.info("Kafka message sent: topic=%s key=%s", topic, key)

    @staticmethod
    def _serialize_key(key: str | None) -> bytes | None:
        if key is None:
            return None
        return key.encode("utf-8")

    @staticmethod
    def _serialize_value(value: dict[str, Any]) -> bytes:
        return json.dumps(value, default=str).encode("utf-8")