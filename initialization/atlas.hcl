variable "schema" {
  type    = string
  default = getenv("BACKEND_DB_SCHEMA")
}

variable "user" {
  type    = string
  default = getenv("BACKEND_DB_USER")
}

variable "password" {
  type    = string
  default = getenv("BACKEND_DB_PASSWORD")
}

data "template_dir" "migrations" {
  path = "migrations-template"
  vars = {
    Schema   = var.schema
    User     = var.user
    Password = var.password
  }
}

env "local" {
  url = "${getenv("POSTGRES_URL")}&search_path=public"

  migration {
    dir = data.template_dir.migrations.url
  }
}