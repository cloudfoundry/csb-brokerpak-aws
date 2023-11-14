variable "name" { type = string }
variable "host" { type = string }
variable "password" {
  type      = string
  sensitive = true
}
variable "tls_port" { type = number }
variable "reader_endpoint" { type = string }