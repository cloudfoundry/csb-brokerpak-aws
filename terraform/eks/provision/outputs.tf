output "cluster_name" {
  value = aws_eks_cluster.eks.name
}

output "endpoint" {
  value = aws_eks_cluster.eks.endpoint
}

output "certificate_authority_data" {
  value = aws_eks_cluster.eks.certificate_authority[0].data
}

output "region" {
  value = var.region
}

output "kubernetes_version" {
  value = aws_eks_cluster.eks.version
}

output "node_group_name" {
  value = aws_eks_node_group.nodes.node_group_name
}

output "node_count" {
  value = var.node_count
}

output "instance_type" {
  value = var.instance_type
}

output "vpc_id" {
  value = aws_vpc.eks.id
}

output "private_subnet_ids_json" {
  value = jsonencode(aws_subnet.private[*].id)
}

output "public_access_cidrs_json" {
  value = jsonencode(local.public_access_cidrs)
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}

output "normalized_instance_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "aws"
    provisioner_family = "aws_eks"
    service            = "kubernetes"
    cluster = {
      name                       = aws_eks_cluster.eks.name
      endpoint                   = aws_eks_cluster.eks.endpoint
      certificate_authority_data = aws_eks_cluster.eks.certificate_authority[0].data
      kubernetes_version         = aws_eks_cluster.eks.version
      region                     = var.region
    }
    topology = {
      control_plane = "managed"
      node_group    = aws_eks_node_group.nodes.node_group_name
      node_count    = var.node_count
      instance_type = var.instance_type
      private_nodes = true
      subnet_ids    = aws_subnet.private[*].id
      vpc_id        = aws_vpc.eks.id
    }
    security = {
      control_plane_logs      = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
      encrypted_node_disk     = true
      endpoint_private_access = true
      endpoint_public_access  = true
      public_access_cidrs     = local.public_access_cidrs
      ssh_enabled             = false
    }
    lifecycle = {
      ttl_hours      = var.ttl_hours
      ttl_expires_at = local.ttl_expires_at
      enforcement    = "external"
    }
  })
}
