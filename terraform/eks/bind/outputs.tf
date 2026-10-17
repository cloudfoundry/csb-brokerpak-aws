output "cluster_name" {
  value = var.cluster_name
}

output "endpoint" {
  value = var.endpoint
}

output "certificate_authority_data" {
  value = var.certificate_authority_data
}

output "region" {
  value = var.region
}

output "kubernetes_version" {
  value = var.kubernetes_version
}

output "ttl_expires_at" {
  value = var.ttl_expires_at
}

output "access_notice" {
  value = local.access_notice
}

output "exec_credential_json" {
  value = jsonencode(local.exec_credential)
}

output "normalized_binding_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "aws"
    provisioner_family = "aws_eks"
    connection_type    = "kubernetes_api"
    endpoint = {
      url                        = var.endpoint
      region                     = var.region
      certificate_authority_data = var.certificate_authority_data
    }
    cluster = {
      name               = var.cluster_name
      kubernetes_version = var.kubernetes_version
    }
    access = {
      mode             = "exec"
      authorization    = "external"
      namespace        = null
      namespace_scoped = false
      notice           = local.access_notice
    }
    credential = {
      format     = "kubernetes_exec"
      exec       = local.exec_credential
      inline     = null
      secret_ref = null
    }
    lifecycle = {
      ttl_expires_at = var.ttl_expires_at
    }
  })
}
