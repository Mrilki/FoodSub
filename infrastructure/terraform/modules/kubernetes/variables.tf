variable "db_username" {
  description = "PostgreSQL username"
  type        = string
  default     = "user"
}

variable "db_password" {
  description = "PostgreSQL password"
  type        = string
  sensitive   = true
}

variable "mongo_uri" {
  description = "MongoDB connection URI"
  type        = string
  default     = "mongodb://mongo:27017"
}

variable "redis_addr" {
  description = "Redis address"
  type        = string
  default     = "redis:6379"
}

variable "kafka_brokers" {
  description = "Kafka bootstrap servers"
  type        = string
  default     = "kafka:9092"
}

variable "kafka_password" {
  description = "Kafka password"
  type        = string
  sensitive   = true
}

variable "auth_service_role_arn" {
  description = "AWS IAM role ARN for auth service"
  type        = string
  default     = ""
}