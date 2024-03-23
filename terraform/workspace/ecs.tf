locals {
  ecs_cluster_name = var.name
  ecr_repo_name    = var.name
  firehose_stream  = local.firehose.name

  services = {
    cpu    = 2048
    memory = 4096

    fluent-bit = {
      service_name = "fluent-bit"
      cpu          = 512
      memory       = 1024
    }

    rest-api = {
      service_name = "rest-api"
      cpu          = 512
      memory       = 1024
      port         = 3000
    }

    redis = {
      service_name = "redis"
      cpu          = 512
      memory       = 1024
      port         = 6379
    }
  }
}

resource "aws_security_group" "microservice" {
  name_prefix = var.name
  vpc_id      = data.terraform_remote_state.core-services.outputs.vpc_main_id

  lifecycle {
    create_before_destroy = true
  }
}

module "ecr" {
  source = "terraform-aws-modules/ecr/aws"

  repository_name = local.ecr_repo_name

  # TODO create machine user to do this
  # repository_read_write_access_arns = ["arn:aws:iam::533267214646:role/AWSFullAccess"]
  repository_lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        description  = "Keep last 10 images",
        selection = {
          tagStatus     = "tagged",
          tagPrefixList = ["v"],
          countType     = "imageCountMoreThan",
          countNumber   = 10
        },
        action = {
          type = "expire"
        }
      }
    ]
  })
}

data "aws_ecr_image" "backend-image" {
  repository_name = local.ecr_repo_name
  image_tag       = "latest"
}

module "ecs" {
  source = "terraform-aws-modules/ecs/aws"

  cluster_name = local.ecs_cluster_name
  cluster_configuration = {
    execute_command_configuration = {
      logging = "OVERRIDE"
      log_configuration = {
        cloud_watch_log_group_name = "/aws/ecs/aws-ecs/${local.ecs_cluster_name}"
      }
    }
  }

  fargate_capacity_providers = {
    FARGATE = {
      default_capacity_provider_strategy = {
        weight = 50
      }
    }

    FARGATE_SPOT = {
      default_capacity_provider_strategy = {
        weight = 50
      }
    }
  }

  services = {
    microservice = {
      cpu    = local.services.cpu
      memory = local.services.memory

      # Container definition(s)
      container_definitions = {
        fluent-bit = {
          cpu       = local.services.fluent-bit.cpu
          memory    = local.services.fluent-bit.memory
          essential = true
          image     = "public.ecr.aws/aws-observability/aws-for-fluent-bit:stable"
          firelens_configuration = {
            type = "fluentbit"
          }
          memory_reservation = 50
        }

        rest-api = {
          cpu       = local.services.rest-api.cpu
          memory    = local.services.rest-api.memory
          essential = true
          image     = data.aws_ecr_image.backend-image.image_uri
          port_mappings = [
            {
              name          = local.services.rest-api.service_name
              containerPort = local.services.rest-api.port
              protocol      = "tcp"
            }
          ]

          # Example image used requires access to write to root filesystem
          readonly_root_filesystem = false

          dependencies = [
            {
              containerName = local.services.fluent-bit.service_name
              condition     = "START"
            },
            {
              containerName = local.services.redis.service_name
              condition     = "START"
            }
          ]

          # Don't use the default AWSLogs driver
          enable_cloudwatch_logging = false

          # Instead let's use awsfirelens and throw everything to firehose
          log_configuration = {
            logDriver = "awsfirelens"
            options = {
              Name                    = "firehose"
              region                  = var.aws_region
              delivery_stream         = local.firehose_stream
              log-driver-buffer-limit = "2097152"
            }
          }
          memory_reservation = 100
        }
      }

      redis = {
        cpu       = local.services.redis.cpu
        memory    = local.services.redis.memory
        essential = true
        image     = "redis:latest"
        port_mappings = [
          {
            name          = local.services.redis.service_name
            containerPort = local.services.redis.port
            protocol      = "tcp"
          }
        ]
      }


      ## Connect to other services of the same cluster
      # service_connect_configuration = {
      #   namespace = "microservice"
      #   service = {
      #     client_alias = {
      #       port     = local.services.backend.port
      #       dns_name = local.services.backend.service_name
      #     }
      #     port_name      = local.services.backend.service_name
      #     discovery_name = local.services.backend.service_name
      #   }
      # }

      load_balancer = {
        service = {
          target_group_arn = aws_lb_target_group.microservice-target-group.arn
          container_name   = local.services.rest-api.service_name
          container_port   = local.services.rest-api.port
        }
      }

      # Task roles
      tasks_iam_role_name        = "${var.name}-tasks"
      tasks_iam_role_description = "Example tasks IAM role for ${var.name}"
      tasks_iam_role_policies    = { TaskExecutionPolicy = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy" }
      tasks_iam_role_statements = [
        {
          actions   = ["firehose:PutRecordBatch"]
          resources = ["arn:aws:firehose:${var.aws_region}:${local.account_id[terraform.workspace]}:deliverystream/${var.name}-firehose-stream"]
        }
      ]


      # Task execution roles
      task_exec_iam_role_name        = "${var.name}-tasks"
      task_exec_iam_role_description = "Task execution IAM role for ${var.name}"
      task_exec_iam_role_policies    = { TaskExecutionPolicy = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy" }
      task_exec_iam_role_statements = [
        {
          actions   = ["firehose:PutRecordBatch"]
          resources = ["arn:aws:firehose:${var.aws_region}:${local.account_id[terraform.workspace]}:deliverystream/${var.name}-firehose-stream"]
        }
      ]

      subnet_ids = data.terraform_remote_state.core-services.outputs.subnet_private_ids
      security_group_rules = {
        ingress = {
          type                     = "ingress"
          from_port                = local.services.rest-api.port
          to_port                  = local.services.rest-api.port
          protocol                 = "tcp"
          description              = "Service Port for ${local.services.rest-api.service_name}"
          source_security_group_id = aws_security_group.microservice.id
        }
        egress = {
          type        = "egress"
          from_port   = 0
          to_port     = 0
          protocol    = "-1"
          cidr_blocks = ["0.0.0.0/0"]
        }
      }
    }
  }
}
