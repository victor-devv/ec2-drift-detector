# Create a VPC
resource "aws_vpc" "sample_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "LocalStack VPC"
    Env  = "dev"
  }
}

resource "aws_subnet" "sample_subnet" {
  vpc_id            = aws_vpc.sample_vpc.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "eu-north-1a"

  tags = {
    Name = "LocalStack Subnet"
    Env  = "dev"
  }
}

resource "aws_security_group" "sample_sg" {
  name        = "localstack-sg"
  description = "Allow SSH and HTTP traffic"
  vpc_id      = aws_vpc.sample_vpc.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "LocalStack Security Group"
    Env  = "dev"
  }
}

resource "aws_instance" "sample_instance" {
  ami                    = "ami-12345678"
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.sample_subnet.id
  vpc_security_group_ids = [aws_security_group.sample_sg.id]

  tags = {
    Name = "LocalStack EC2"
    Env  = "dev"
  }
}
