resource "random_string" "username" {
  length  = 16
  special = false
  numeric = false
}

resource "random_password" "password" {
  length           = 64
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "csbsqlserver_binding" "binding" {
  username = random_string.username.result
  password = random_password.password.result
  roles    = ["db_ddladmin", "db_datareader", "db_datawriter", "db_accessadmin"]
}

