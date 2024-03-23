variable "aws_region" {
  default     = "us-east-1"
  description = "AWS region"
  type        = string
}

variable "name" {
  default     = "microservice"
  description = "Name of the microservice"
  type        = string
}

variable "health_url" {
  default     = "/health"
  description = "Endpoint for Health Url"
  type        = string
}

variable "application_port" {
  default     = 3000
  description = "Endpoint where the application will be running"
  type        = number
}

variable "alb_rule_priority" {
  default     = 1000
  description = "Priority for this particular project. They all must be different."
  type        = number
}
