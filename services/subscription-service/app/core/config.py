from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    database_url: str

    model_config = SettingsConfigDict(
        env_file=".env",
        extra="ignore",
    )

    kafka_bootstrap_servers: str = "localhost:9092"
    kafka_group_id: str = "subscription-service"


settings = Settings()