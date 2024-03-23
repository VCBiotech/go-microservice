terraform {
  backend "s3" {
    bucket               = "vcbiotech-terraform-state"
    key                  = "terraform"
    region               = "us-east-1"
    encrypt              = true
    kms_key_id           = "8c8bcdb0-8bbd-4490-88a0-71bc38627c3d"
    dynamodb_table       = "vcbiotech-terraform-state-lock"
    workspace_key_prefix = "microservice"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.36.0"
    }
  }

  required_version = ">= 1.7.3"
}
