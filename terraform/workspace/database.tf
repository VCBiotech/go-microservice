locals {
  database_desired_count = {
    prod = 2
    test = 1
  }

  database_serverless_capacity = {
    prod = {
      min = 0.5
      max = 2.0
    }
    test = {
      min = 0.5
      max = 1.0
    }
  }

  database_name = "core"
  cluster_name  = "core"
}

module "database" {
  source = "git@github.com:vcbiotech/infrastructure.git//terraform/modules/aurora-serverless?ref=main"

  cidr_blocks   = ["0.0.0.0/0"]
  cluster_name  = local.cluster_name
  database_name = local.database_name
  subnet_ids    = data.terraform_remote_state.core-services.outputs.subnet_private_ids
  vpc_id        = data.terraform_remote_state.core-services.outputs.vpc_main_id
  minimum_acu   = local.database_serverless_capacity[terraform.workspace].min
  maximum_acu   = local.database_serverless_capacity[terraform.workspace].max
}
