#!/bin/bash
set -e

echo "==> Initializing LocalStack with test AWS resources..."

# Configure AWS CLI to use LocalStack
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export ENDPOINT="--endpoint-url=http://localhost:4566"

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
awslocal sts get-caller-identity

# Create VPC
echo "Creating VPC..."
VPC_ID=$(awslocal ec2 create-vpc --cidr-block 10.0.0.0/16 --query 'Vpc.VpcId' --output text)
awslocal ec2 create-tags --resources $VPC_ID --tags Key=Name,Value=TestVPC

# Create Subnet
echo "Creating Subnet..."
SUBNET_ID=$(awslocal ec2 create-subnet --vpc-id $VPC_ID --cidr-block 10.0.1.0/24 --query 'Subnet.SubnetId' --output text)
awslocal ec2 create-tags --resources $SUBNET_ID --tags Key=Name,Value=TestSubnet

# Create Internet Gateway
echo "Creating Internet Gateway..."
IGW_ID=$(awslocal ec2 create-internet-gateway --query 'InternetGateway.InternetGatewayId' --output text)
awslocal ec2 attach-internet-gateway --internet-gateway-id $IGW_ID --vpc-id $VPC_ID

# Create Route Table
echo "Creating Route Table..."
ROUTE_TABLE_ID=$(awslocal ec2 create-route-table --vpc-id $VPC_ID --query 'RouteTable.RouteTableId' --output text)
awslocal ec2 create-route --route-table-id $ROUTE_TABLE_ID --destination-cidr-block 0.0.0.0/0 --gateway-id $IGW_ID
awslocal ec2 associate-route-table --route-table-id $ROUTE_TABLE_ID --subnet-id $SUBNET_ID

# Create Security Group
echo "Creating Security Group..."
SG_ID=$(awslocal ec2 create-security-group --group-name TestSG --description "Test Security Group" --vpc-id $VPC_ID --query 'GroupId' --output text)
awslocal ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 22 --cidr 0.0.0.0/0
awslocal ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 80 --cidr 0.0.0.0/0
awslocal ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 443 --cidr 0.0.0.0/0

# Create Key Pair
echo "Creating Key Pair..."
awslocal ec2 create-key-pair --key-name TestKey --query 'KeyMaterial' --output text > /tmp/TestKey.pem
chmod 400 /tmp/TestKey.pem

# Create an AMI for testing
echo "Creating AMI..."
AMI_ID=$(awslocal ec2 create-image --name TestAMI --description "Test AMI" --no-reboot --query 'ImageId' --output text || echo ami-12345678)

# Create EC2 Instances
echo "Creating EC2 Instances..."

# Create instance 1 (t2.micro)
INSTANCE_1_ID=$(awslocal ec2 run-instances \
    --image-id $AMI_ID \
    --count 1 \
    --instance-type t2.micro \
    --key-name TestKey \
    --security-group-ids $SG_ID \
    --subnet-id $SUBNET_ID \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=TestInstance1},{Key=Environment,Value=Test},{Key=Project,Value=DriftDetector}]' \
    --query 'Instances[0].InstanceId' \
    --output text)

echo "Created Instance 1: $INSTANCE_1_ID"

# Create instance 2 (t2.small)
INSTANCE_2_ID=$(awslocal ec2 run-instances \
    --image-id $AMI_ID \
    --count 1 \
    --instance-type t2.small \
    --key-name TestKey \
    --security-group-ids $SG_ID \
    --subnet-id $SUBNET_ID \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=TestInstance2},{Key=Environment,Value=Production},{Key=Project,Value=DriftDetector}]' \
    --query 'Instances[0].InstanceId' \
    --output text)

echo "Created Instance 2: $INSTANCE_2_ID"

# Create a volume and attach it to instance 1
echo "Creating and attaching EBS volume..."
VOLUME_ID=$(awslocal ec2 create-volume \
    --availability-zone us-west-2a \
    --size 10 \
    --volume-type gp2 \
    --tag-specifications 'ResourceType=volume,Tags=[{Key=Name,Value=TestVolume},{Key=Environment,Value=Test}]' \
    --query 'VolumeId' \
    --output text)

awslocal ec2 attach-volume \
    --volume-id $VOLUME_ID \
    --instance-id $INSTANCE_1_ID \
    --device /dev/sdf

echo "Attached volume $VOLUME_ID to instance $INSTANCE_1_ID"

# Save instance IDs to a file for reference
echo "Saving instance IDs to file..."
echo "INSTANCE_1_ID=$INSTANCE_1_ID" > /tmp/instance_ids.env
echo "INSTANCE_2_ID=$INSTANCE_2_ID" >> /tmp/instance_ids.env

echo "==> LocalStack initialization complete!"
