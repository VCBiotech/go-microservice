variable "aws_region" {
  default     = "us-east-1"
  description = "AWS region"
  type        = string
}

variable "name" {
  default     = "equilibria-files"
  description = "Name of the microservice"
  type        = string
}

variable "ecr_name" {
  default     = "vcbiotech/equilibria-files"
  description = "Name of the image ECR repo"
  type        = string
}

variable "gh_repo" {
  default     = "VCBiotech/equilibria-files"
  description = "Name of the Github repo"
  type        = string
}

variable "docker_tag" {
  default     = "latest"
  description = "Image versioning system based of SHA"
  type        = string
}

variable "allowed_ip_cidrs" {
  default     = null
  description = "List of CIDRs allowed to access this service. Leave blank to allow all"
  type        = list(string)
}

variable "health_url" {
  default     = "/api/health"
  description = "Endpoint for Health Url"
  type        = string
}

variable "application_port" {
  default     = 3000
  description = "Endpoint where the application will be running"
  type        = number
}

variable "alb_rule_priority" {
  default     = 1002
  description = "Priority for this particular project. They all must be different."
  type        = number
}

variable "environment" {
  default     = "test"
  description = "Current environment"
  type        = string
}

variable "domain" {
  default     = "equilibriahrt.com"
  type        = string
  description = "The domain name to use for the certificate"
}

variable "subdomain" {
  default     = "files"
  type        = string
  description = "The subdomain name to use for the certificate"
}

variable "github_token" {
  type        = string
  description = "GitHub token"
  default     = ""
}

variable "management_account_id" {
  type        = string
  default     = "224298540768"
  description = "Management account ID"
}

variable "account_id" {
  type        = string
  description = "Account ID"
  default     = "533267214646"
}

variable "clerk_publishable_key" {
  type        = string
  description = "Clerk publishable key"
  default     = "pk_live_Y2xlcmsuZXF1aWxpYnJpYWhydC5jb20k"
}
