provider "aws" {
  region = var.region
}

provider "csbmajorengineversion" {
  region = var.region
  engine = var.engine
}
