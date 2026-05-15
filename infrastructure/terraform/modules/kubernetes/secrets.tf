
resource "kubernetes_secret" "jwt_keys" {
  metadata {
    name      = "jwt-keys"
    namespace = kubernetes_namespace.food_subscription.metadata[0].name
  }
  data = {
    "private.pem" = file("${path.module}/secrets/private.pem")
    "public.pem"  = file("${path.module}/secrets/public.pem")
  }
  type = "Opaque"
}

resource "kubernetes_secret" "database_credentials" {
  metadata {
    name      = "db-credentials"
    namespace = kubernetes_namespace.food_subscription.metadata[0].name
  }
  data = {
    "postgres-user"     = base64encode(var.db_username)
    "postgres-password" = base64encode(var.db_password)
    "mongo-uri"         = base64encode(var.mongo_uri)
    "redis-addr"        = base64encode(var.redis_addr)
    "kafka-brokers"     = base64encode(var.kafka_brokers)
  }
  type = "Opaque"
}

resource "kubernetes_secret" "kafka_credentials" {
  metadata {
    name      = "kafka-credentials"
    namespace = "kafka"
  }
  data = {
    "username" = base64encode("kafka-user")
    "password" = base64encode(var.kafka_password)
  }
  type = "Opaque"
}