terraform {
  backend "s3" {
    bucket       = "vcbiotech-infrastructure-state"
    key          = "core-services/equilibria-file-manager/test/terraform"
    region       = "us-east-1"
    encrypt      = true
    kms_key_id   = "ddff0cbe-6ce5-4419-ae52-dea807798d60"
    use_lockfile = true
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.36.0"
    }
  }

  required_version = ">= 1.5"
}
