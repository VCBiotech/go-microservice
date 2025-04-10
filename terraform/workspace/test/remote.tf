data "terraform_remote_state" "core_services" {
  backend = "s3"
  config = {
    bucket       = "vcbiotech-infrastructure-state"
    key          = "core-services/test/terraform"
    region       = "us-east-1"
    encrypt      = true
    kms_key_id   = "ddff0cbe-6ce5-4419-ae52-dea807798d60"
    use_lockfile = true
  }
}
