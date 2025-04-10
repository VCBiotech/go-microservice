provider "aws" {
  alias = "route53"
  assume_role {
    role_arn = "arn:aws:iam::${var.management_account_id}:role/CertificateManager"
  }
  region = var.aws_region
}

provider "aws" {
  assume_role {
    role_arn = "arn:aws:iam::${var.account_id}:role/FullAWSAccess"
  }
  default_tags {
    tags = {
      "Owner" = "Terraform"
    }
  }
  region = var.aws_region
}

provider "aws" {
  alias = "acm"
  assume_role {
    role_arn = "arn:aws:iam::${var.account_id}:role/FullAWSAccess"
  }
  default_tags {
    tags = {
      "Owner" = "Terraform"
    }
  }
  region = var.aws_region
}
