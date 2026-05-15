from locust import HttpUser, task, between, events
import random
import json

class RegularUser(HttpUser):
    """Симуляция обычного пользователя (просмотр меню, подписка)"""
    wait_time = between(1, 3)  # Пауза между запросами 1-3 секунды

    @task(5)
    def get_menu(self):
        """Просмотр меню (самое частое действие)"""
        self.client.get("/api/v1/menu", name="GET /menu")

    @task(3)
    def get_menu_by_id(self):
        """Просмотр конкретного блюда"""
        menu_ids = [
            "65f5a1b2c3d4e5f6a7b8c9d0",
            "65f5a1b2c3d4e5f6a7b8c9d1",
            "65f5a1b2c3d4e5f6a7b8c9d2"
        ]
        menu_id = random.choice(menu_ids)
        self.client.get(f"/api/v1/menu/{menu_id}", name="GET /menu/:id")

    @task(2)
    def search_menu(self):
        """Поиск блюд по тегам"""
        tags = random.choice([
            "итальянская",
            "веган",
            "быстро",
            "здоровое",
            "популярное"
        ])
        self.client.get(f"/api/v1/menu/search?tags={tags}", name="GET /menu/search")

    @task(1)
    def health_check(self):
        """Health check endpoint"""
        self.client.get("/health", name="GET /health")

    def on_start(self):
        """Вызывается один раз при старте пользователя"""
        pass



class AdminUser(HttpUser):
    """Симуляция администратора (CRUD операции)"""
    wait_time = between(5, 10)

    def on_start(self):
        """Логин как админ при старте"""
        response = self.client.post("/api/v1/auth/login",
            json={
                "email": "admin@food-subscription.local",
                "password": "admin123"
            }
        )
        if response.status_code == 200:
            self.token = response.json().get("access_token")
            self.headers = {"Authorization": f"Bearer {self.token}"}
        else:
            self.token = None
            self.headers = {}

    @task(3)
    def create_menu_item(self):
        """Создание нового блюда"""
        if self.token:
            self.client.post("/api/v1/admin/menu",
                json={
                    "name": f"Тестовое блюдо {random.randint(1, 10000)}",
                    "description": "Тестовое описание",
                    "category": random.choice(["Обед", "Ужин", "Завтрак"]),
                    "ingredients": ["Ингредиент 1", "Ингредиент 2"],
                    "kbju": {
                        "calories": random.randint(100, 800),
                        "proteins": random.randint(5, 50),
                        "fats": random.randint(5, 50),
                        "carbs": random.randint(10, 100)
                    },
                    "tags": ["тест", "автоматически"],
                    "image_url": "https://example.com/test.jpg",
                    "is_available": True
                },
                headers=self.headers,
                name="POST /admin/menu"
            )

    @task(1)
    def list_archives(self):
        """Просмотр архивов"""
        if self.token:
            self.client.get("/api/v1/admin/archive",
                headers=self.headers,
                name="GET /admin/archive"
            )


class KafkaProducerUser(HttpUser):
    """Имитация Python сервиса который отправляет события в Kafka"""
    wait_time = between(10, 30)

    @task(1)
    def send_order_event(self):
        """Симуляция отправки order.scheduled события"""

        self.client.post("/api/v1/orders",
            json={
                "subscription_id": f"sub_{random.randint(1, 1000)}",
                "delivery_date": "2024-03-20T10:00:00Z",
                "menu_item_ids": [
                    "65f5a1b2c3d4e5f6a7b8c9d0",
                    "65f5a1b2c3d4e5f6a7b8c9d1"
                ]
            },
            name="POST /orders (Kafka event simulation)"
        )


@events.request.add_listener
def on_request(request_type, name, response_time, response_length, response, context, exception, **kwargs):
    """Логирование каждого запроса"""
    if exception:
        print(f"ERROR: {name} - {exception}")
    elif response_time > 500:
        print(f"SLOW: {name} took {response_time}ms")


@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """При старте теста"""
    print("Locust test starting...")
    print(f"Target: {environment.host}")


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """При остановке теста"""
    print("Locust test completed!")
    print(f"Total requests: {environment.stats.total.num_requests}")
    print(f"Failed requests: {environment.stats.total.num_failures}")