resource "aws_lb_listener_rule" "microservice_listener_rule" {
  action {
    target_group_arn = aws_lb_target_group.microservice_target_group.arn
    type             = "forward"
  }

  condition {
    host_header {
      values = ["${var.subdomain}.${var.domain}"]
    }
  }

  # Block cidrs directly
  dynamic "condition" {
    for_each = var.allowed_ip_cidrs == null ? [] : [1]
    content {
      source_ip {
        values = var.allowed_ip_cidrs
      }
    }
  }

  listener_arn = data.terraform_remote_state.core_services.outputs.alb_public_listener_https_arn
  priority     = var.alb_rule_priority
}

resource "aws_lb_target_group" "microservice_target_group" {
  name = "${var.name}-target-group"

  deregistration_delay = 30

  health_check {
    healthy_threshold   = 2
    interval            = 10
    path                = var.health_url
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 5
  }

  port        = var.application_port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = data.terraform_remote_state.core_services.outputs.vpc_main_id
}
