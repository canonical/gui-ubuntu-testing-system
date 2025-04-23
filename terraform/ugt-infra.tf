terraform {
  required_providers {
    juju = {
      version = "~> 0.10.1"
      source  = "juju/juju"
    }
  }
}

provider "juju" {}

# variable "dummy_variable" {
#   description = "My dummy variable"
#   type        = string
#   default     = "dummy"
# }

locals {
  juju_model = "asdf"
}


resource "juju_application" "postgresql" {
  name  = "postgresql"
  model = local.juju_model
  trust = true

  constraints = "mem=8G"

  storage_directives = {
    pgdata = "10G"  # TBD!
  }

  charm {
    name     = "postgresql"
    channel  = "14/stable"
    base     = "ubuntu@22.04"
  }
}

resource "juju_application" "handler" {
  name  = "handler"
  model = local.juju_model

  constraints = "mem=8G"

  charm {
    name     = "ubuntu-gui-testing-handler"
    base     = "ubuntu@24.04"
  }
}

resource "juju_integration" "handler-postgresql" {
  model = local.juju_model

  application {
    name = juju_application.handler.name
    endpoint = "db"
  }

  application {
    name = juju_application.postgresql.name
    endpoint = "db-admin"
  }
}

################################################
# Notes down here
# juju integrate db ubuntu-gui-testing-handler
# failed with
# ERROR ambiguous relation: "db ubuntu-gui-testing-handler" could refer to "ubuntu-gui-testing-handler:db db:db"; "ubuntu-gui-testing-handler:db db:db-admin"
# juju integrate db:db ubuntu-gui-testing-handler:db
# works, but it may require ubuntu-gui-testing-handler:db-admin, especially if the handler handles the schema

