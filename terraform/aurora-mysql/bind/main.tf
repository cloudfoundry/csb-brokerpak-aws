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

resource "mysql_user" "newuser" {
  user               = random_string.username.result
  plaintext_password = random_password.password.result
  host               = "%"
  tls_option         = "NONE"
}

resource "mysql_grant" "newuser" {
  user       = mysql_user.newuser.user
  database   = var.name
  host       = mysql_user.newuser.host
  privileges = ["ALL"]
}