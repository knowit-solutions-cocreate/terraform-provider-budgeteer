terraform {
  required_providers {
    budgeteer = {
      source  = "hashicorp.com/dev/budgeteer"
      version = "1.0.0"
    }
  }
}

provider "budgeteer" {
  host = "http://localhost:8080"
  api_key = "sk-XXXXXX"
  # host              = var.budgeteer_host
  # api_key           = var.budgeteer_api_key
}

resource "budgeteer_api_key" "example" {
  name   = "example-key"
  budget = 100
}

resource "budgeteer_api_key" "eriks_key" {
  name   = "eriks-key"
  budget = 10000
}

resource "budgeteer_api_key" "jimmys_key" {
  name   = "jimmys-key"
  budget = 10000
}