locals {
  account_id = {
    "prod" = 339713108680
    "test" = 533267214646
  }
}

data "aws_caller_identity" "current" {}
