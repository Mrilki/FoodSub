terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

module "kubernetes" {
  source = "../../modules/kubernetes"

  db_username   = "user"
  db_password   = "pass"
  mongo_uri     = "mongodb://mongo:27017"
  redis_addr    = "redis:6379"
  kafka_brokers = "kafka:9092"
  kafka_password = "kafka-secret"
}

output "namespace_name" {
  value = module.kubernetes.kubernetes_namespace.food_subscription.metadata[0].name
}