variable "instance_name" { type = string }
variable "region" { type = string }
variable "ttl_hours" {
  type = number

  validation {
    condition     = var.ttl_hours == 8
    error_message = "ttl_hours must be exactly 8 for the sandbox plan."
  }
}
variable "node_count" {
  type = number

  validation {
    condition     = var.node_count == 3
    error_message = "node_count must be exactly 3 for the sandbox plan."
  }
}
variable "kubernetes_version" {
  type = string

  validation {
    condition     = can(regex("^1\\.[0-9]+$", var.kubernetes_version))
    error_message = "kubernetes_version must be an EKS minor version such as 1.34."
  }
}
variable "instance_type" { type = string }
variable "public_access_cidrs_json" {
  type = string

  validation {
    condition = can([
      for cidr in jsondecode(var.public_access_cidrs_json) : cidrhost(cidr, 0)
    ]) && length(jsondecode(var.public_access_cidrs_json)) > 0
    error_message = "public_access_cidrs_json must be a non-empty JSON array of valid CIDRs."
  }
}
variable "labels" { type = map(any) }
