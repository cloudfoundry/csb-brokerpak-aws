locals {
  jdbc_tls_string = (var.require_ssl ? format("encrypt=true;trustServerCertificate=false;hostNameInCertificate=%s", var.hostname) : "encrypt=false")
  uri_tls_string  = (var.require_ssl ? format("encrypt=true&TrustServerCertificate=false&HostNameInCertificate=%s", var.hostname) : "encrypt=false")
}