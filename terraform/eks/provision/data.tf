data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  availability_zones  = slice(data.aws_availability_zones.available.names, 0, 2)
  public_access_cidrs = jsondecode(var.public_access_cidrs_json)
  ttl_expires_at      = timeadd(plantimestamp(), "${var.ttl_hours}h")
  node_group_name     = "${var.instance_name}-nodes"

  common_tags = merge(var.labels, {
    ManagedBy   = "cloud-service-broker"
    Environment = "sandbox"
    TTLExpiry   = local.ttl_expires_at
  })
}
