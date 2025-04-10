locals {
  services = {
    cpu    = 512
    memory = 1024

    files = {
      name   = "files"
      cpu    = 256
      memory = 512
      port   = 3000
    }
  }
}

# Used to store docker images
module "ecr" {
  source = "terraform-aws-modules/ecr/aws"

  repository_name                 = var.ecr_name
  repository_image_tag_mutability = "MUTABLE"

  repository_lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        description  = "Keep last 30 images",
        selection = {
          tagStatus   = "any",
          countType   = "imageCountMoreThan",
          countNumber = 30
        },
        action = {
          type = "expire"
        }
      }
    ]
  })
}

module "ecs" {
  source = "terraform-aws-modules/ecs/aws//modules/service"

  name        = var.name
  cluster_arn = data.terraform_remote_state.core_services.outputs.cluster_arn

  # Scaling
  platform_version                   = "LATEST"
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100
  desired_count                      = 1

  # Specs
  cpu    = local.services.cpu
  memory = local.services.memory

  # Runtime for containers
  runtime_platform = {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  # This will force the service to always re-deploy tasks.  We want this so that
  # new tasks are deployed if we decide update/reuse an existing docker tag which
  # happens to already be running.
  force_new_deployment = true

  # This will instruct terraform to wait until the ECS task deployment either
  # finishes successfully or fails
  wait_for_steady_state = true

  # This block instructs ECS to rollback to a previously working task if the new
  # deployment fails to reach a steady state (instead of indefinitely retrying).
  deployment_circuit_breaker = {
    enable   = true
    rollback = true
  }

  # Networking and Firewalls 
  create_security_group = false
  security_group_ids    = [aws_security_group.equilibria_app_sg.id]

  # Subnet to place tasks
  subnet_ids = data.terraform_remote_state.core_services.outputs.subnet_private_ids

  # Container definition(s)
  container_definitions = [
    # Main Service
    {
      name      = local.services.files.name
      cpu       = local.services.files.cpu
      memory    = local.services.files.memory
      essential = true
      image     = "${module.ecr.repository_url}:${var.docker_tag}"

      # Store logs
      cloudwatch_log_group_name = "/aws/ecs/${var.name}"

      // Add environment variables
      environment = [
        {
          name  = "APP_URL"
          value = "https://${var.subdomain}.${var.domain}"
        },
        {
          name  = "APP_ENV"
          value = var.environment
        },
        {
          name  = "CLERK_SECRET_KEY"
          value = local.clerk_secret_key
        }
      ]

      // Add secrets
      secrets = [
        {
          name      = "SECRETS"
          valueFrom = aws_secretsmanager_secret.secrets.arn,
        }
      ]

      port_mappings = [
        {
          name          = "${var.name}"
          containerPort = var.application_port
          protocol      = "tcp"
        }
      ]

      tags = {
        Terraform = "true"
      }

    }
  ]

  # Task roles which is the one the actual tasks use
  tasks_iam_role_name        = "${var.name}-tasks"
  tasks_iam_role_description = "IAM Task Role for ${var.name}"
  tasks_iam_role_statements = [
    {
      sid       = "AllowGetSecrets"
      actions   = ["secretsmanager:GetSecretValue"]
      resources = [aws_secretsmanager_secret.secrets.arn]
    }
  ]

  # Load balancer configuration
  load_balancer = {
    service = {
      target_group_arn = aws_lb_target_group.microservice_target_group.arn
      container_name   = local.services.files.name
      container_port   = local.services.files.port
    }
  }
}

resource "aws_security_group" "equilibria_app_sg" {
  name        = "${var.name}-security-group"
  description = "Security group for ${var.name} service."
  vpc_id      = data.terraform_remote_state.core_services.outputs.vpc_main_id

  ingress {
    from_port       = local.services.files.port
    to_port         = local.services.files.port
    protocol        = "tcp"
    security_groups = [data.terraform_remote_state.core_services.outputs.alb_public_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "-1"
  }
}
