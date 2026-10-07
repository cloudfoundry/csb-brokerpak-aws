locals {
  access_notice = "Metadata-only MVP: this binding creates no Kubernetes namespace, RBAC, EKS access entry, IAM principal, or isolation boundary. The caller must use an independently authorized short-lived AWS identity."

  exec_credential = {
    apiVersion         = "client.authentication.k8s.io/v1beta1"
    command            = "aws"
    args               = ["eks", "get-token", "--cluster-name", var.cluster_name, "--region", var.region]
    interactiveMode    = "Never"
    provideClusterInfo = false
  }
}
