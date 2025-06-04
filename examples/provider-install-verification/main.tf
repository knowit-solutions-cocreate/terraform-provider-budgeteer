terraform {
  required_providers {
    budgeteer = {
      source  = "hashicorp.com/dev/budgeteer"
      version = "1.0.0"
    }
  }
}

provider "budgeteer" {
  host              = var.budgeteer_host
  budgeteer_api_key = var.budgeteer_api_key
}

resource "budgeteer_api_key" "example" {
  name   = "example-key"
  budget = 100
}
