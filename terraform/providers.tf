provider "aws" {

  access_key                  = var.aws_access_key_id
  secret_key                  = var.aws_secret_access_key
  region                      = var.aws_default_region

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  

  endpoints {
    ec2 = var.ec2_endpoint
    iam = var.ec2_endpoint
  }

  insecure = true
}
