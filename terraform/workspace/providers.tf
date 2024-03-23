provider "aws" {
  assume_role {
    role_arn = "arn:aws:iam::${local.account_id[terraform.workspace]}:role/FullAWSAccess"
  }
  default_tags {
    tags = {
      "Owner" = "Terraform"
    }
  }
  region = var.aws_region
}
