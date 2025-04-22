resource "aws_instance" "example" {
  id             = "i-abc123"
  ami            = "ami-abc123"
  instance_type  = "t3.medium"
  subnet_id      = "subnet-xyz"
  vpc_security_group_ids = ["sg-123"]

  tags = {
    env = "production"
  }
}
