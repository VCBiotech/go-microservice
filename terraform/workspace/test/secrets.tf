# Place to store secrets
resource "aws_secretsmanager_secret" "secrets" {
  name_prefix = "${var.name}/environment-variables/"
}

# Get DATABASE_URL from secret which is in JSON format
data "aws_secretsmanager_secret_version" "secrets_version" {
  secret_id = aws_secretsmanager_secret.secrets.id
}

# Parse DATABASE_URL from secret and make it sensitive
locals {
  database_url     = sensitive(jsondecode(data.aws_secretsmanager_secret_version.secrets_version.secret_string)["DATABASE_URL"])
  clerk_secret_key = sensitive(jsondecode(data.aws_secretsmanager_secret_version.secrets_version.secret_string)["CLERK_SECRET_KEY"])
}


