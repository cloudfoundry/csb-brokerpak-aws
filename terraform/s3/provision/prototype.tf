data "terraform_remote_state" "prev_state" {
  backend = "local"
  config = {
    path = "./terraform.tfstate"
  }
  defaults = {
    inputs : var.properties
  }
}

locals {
  deprecated_inputs = []

  default_inputs    = data.terraform_remote_state.prev_state.defaults.inputs
  last_inputs       = data.terraform_remote_state.prev_state.outputs.inputs
  last_valid_inputs = { for k, v in local.last_inputs : k => v if !contains(local.deprecated_inputs, k) }

  inputs            = merge(local.default_inputs, local.last_valid_inputs, var.inputs)
  unsupported_props = join(",", setsubtract(keys(local.inputs), keys(var.properties)))

  invalid_prohibit_updates = join(",", setsubtract(toset(var.prohibit_updates), keys(var.properties)))
}

resource "terraform_data" "strongly_typed_inputs" {
  lifecycle {
    precondition {
      condition     = length(local.unsupported_props) == 0
      error_message = "unsupported properties specified as inputs: ${local.unsupported_props}"
    }

    precondition {
      condition     = length(local.invalid_prohibit_updates) == 0
      error_message = "unsupported properties specified as prohibit_updates: ${local.invalid_prohibit_updates}"
    }
  }
}

resource "terraform_data" "prohibit_update" {
  # Don't run prohibit_update during instance creation
  count = local.last_inputs.ready ? length(var.prohibit_updates) : 0

  lifecycle {
    precondition {
      condition     = local.inputs[var.prohibit_updates[count.index]] == local.last_inputs[var.prohibit_updates[count.index]]
      error_message = "${var.prohibit_updates[count.index]} can't be modified after creation"
    }
  }
}

variable "inputs" {
  type    = any
  default = {}
}

output "inputs" {
  value = merge(local.inputs, { "ready" : true })
}
