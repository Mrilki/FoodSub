terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "~> 1.14"
    }
  }
}

resource "kubernetes_namespace" "food_subscription" {
  metadata {
    name = "food-subscription"
    labels = {
      "istio-injection" = "enabled"
      "monitoring"      = "enabled"
      "app.kubernetes.io/part-of" = "food-subscription"
    }
  }
}

resource "kubernetes_service_account" "auth_service" {
  metadata {
    name      = "auth-service"
    namespace = kubernetes_namespace.food_subscription.metadata[0].name
    annotations = {
      "eks.amazonaws.com/role-arn" = var.auth_service_role_arn
    }
  }
}

resource "kubernetes_service_account" "subscription_service" {
  metadata {
    name      = "subscription-service"
    namespace = kubernetes_namespace.food_subscription.metadata[0].name
  }
}

resource "kubernetes_service_account" "catalog_service" {
  metadata {
    name      = "catalog-service"
    namespace = kubernetes_namespace.food_subscription.metadata[0].name
  }
}