import json
import logging
from collections.abc import Awaitable, Callable
from typing import Any

from aiokafka import AIOKafkaConsumer

logger = logging.getLogger(__name__)

KafkaHandler = Callable[[dict[str, Any]], Awaitable[None]]


class KafkaConsumerService:
    def __init__(
        self,
        bootstrap_servers: str,
        group_id: str,
        topics: list[str],
    ):
        self.bootstrap_servers = bootstrap_servers
        self.group_id = group_id
        self.topics = topics
        self._consumer: AIOKafkaConsumer | None = None
        self._handlers: dict[str, KafkaHandler] = {}

    def register_handler(self, topic: str, handler: KafkaHandler) -> None:
        self._handlers[topic] = handler

    async def start(self) -> None:
        if self._consumer is not None:
            return

        self._consumer = AIOKafkaConsumer(
            *self.topics,
            bootstrap_servers=self.bootstrap_servers,
            group_id=self.group_id,
            enable_auto_commit=False,
            key_deserializer=self._deserialize_key,
            value_deserializer=self._deserialize_value,
        )

        await self._consumer.start()
        logger.info("Kafka consumer started: topics=%s", self.topics)

    async def stop(self) -> None:
        if self._consumer is None:
            return

        await self._consumer.stop()
        self._consumer = None
        logger.info("Kafka consumer stopped")

    async def consume_forever(self) -> None:
        if self._consumer is None:
            raise RuntimeError("Kafka consumer is not started")

        try:
            async for message in self._consumer:
                handler = self._handlers.get(message.topic)

                if handler is None:
                    logger.warning("No handler for Kafka topic: %s", message.topic)
                    await self._consumer.commit()
                    continue

                try:
                    await handler(message.value)
                    await self._consumer.commit()
                except Exception:
                    logger.exception(
                        "Kafka message processing failed: topic=%s partition=%s offset=%s",
                        message.topic,
                        message.partition,
                        message.offset,
                    )
        finally:
            await self.stop()

    @staticmethod
    def _deserialize_key(key: bytes | None) -> str | None:
        if key is None:
            return None
        return key.decode("utf-8")

    @staticmethod
    def _deserialize_value(value: bytes) -> dict[str, Any]:
        return json.loads(value.decode("utf-8"))