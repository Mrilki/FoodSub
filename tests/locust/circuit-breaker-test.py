from locust import HttpUser, task, events
import time

class CircuitBreakerTest(HttpUser):
    wait_time = between(0.1, 0.5)

    @task
    def flood_requests(self):
        """Отправляем много запросов чтобы триггернуть Circuit Breaker"""
        self.client.get("/api/v1/menu", name="Flood /menu")

    @task
    def check_health(self):
        """Проверяем health endpoint"""
        self.client.get("/health", name="Health Check")

@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    print("⚡ Circuit Breaker Test Starting...")
    print("📋 Steps:")
    print("1. Запустить тест на 2 минуты")
    print("2. Остановить catalog-service под")
    print("3. Наблюдать за Circuit Breaker в Istio")
    print("4. Circuit Breaker должен исключить под из пула")
    print("5. Восстановить под")
    print("6. Circuit Breaker должен вернуть под в пул")