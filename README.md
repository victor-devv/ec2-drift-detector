# EC2 Drift Detector

A Go-based CLI tool that detects configuration drift between **AWS EC2 instances** and their corresponding **Terraform definitions**.

---

## 🧠 Project Overview

The EC2 Drift Detector identifies mismatches between the actual state of EC2 instances in AWS and their declared configurations in Terraform. This tool is helpful for:

- Infrastructure audits
- Change detection
- CI/CD integrity checks
- Compliance verification

The application can be configured through:

1. Configuration files (YAML)
2. Environment variables via .env or .envrc
3. Command-line flags

---

## 🚀 Features

- ✅ Compares multiple attributes: `instance_type`, `ami`, `tags`, `security_groups`, and more
- ✅ Supports concurrent and sequential drift detection
- ✅ Outputs results in console or JSON format
- ✅ Modular and testable design
- ✅ Works with local or remote Terraform `.tfstate` and HCL configurations
- ✅ Built-in support for mocking AWS via [LocalStack](https://github.com/localstack/localstack)

---

## 🛠️ Setup & Installation

### 📦 Prerequisites

- [Go 1.24+](https://golang.org/dl/)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Terraform](https://developer.hashicorp.com/terraform/downloads)
- [Direnv](https://direnv.net/)

### 🔧 Build from source

Clone the repository
```bash
git clone https://github.com/victor-devv/ec2-drift-detector.git && cd ec2-drift-detector
```

Build the binary
```bash
make build
```

### ⚙️ Fill configuration values

#### Configuration file

Create config.yaml from sample config file
```bash
cp config.yaml.example config.yaml
```

#### Environment Variables (.envrc or .env)

Create .envrc from sample .envrc file (no need for exports if using .env)
```bash
cp .envrc.example .envrc
```

Load variables into shell (if using .envrc)
```bash
direnv allow
```

### 🐳 Start LocalStack

Start Localstack docker container
```bash
make localstack-up
```

### 🐳 Create AWS resources

Run terraform init, plan and apply

```bash
make tf-init && make tf-plan && make tf-apply
```

### 🧪 Run Drift Detection 

#### Built binary

```bash
make run-binary
```

#### Or using go run

```bash
make run
```

### 🧪 Run Tests

```bash
make test
```

View test coverage report
```bash
make test-report
```

### Running Raw Commands

To check for drift in all instances:

```bash
./drift-detector detect
```

To check for drift in a specific instance:

```bash
./drift-detector detect i-1234567890abcdef0
```

To run as a server with scheduled checks:

```bash
./drift-detector server
```

To view current configuration:

```bash
./drift-detector config show
```

To reload configuration:

```bash
./drift-detector config reload
```

---

## 🧭 CLI Usage & Examples

### ⚙️ Command-line flags for detect command

| Flag                | Type      | Default     | Description                                      |
|----------------     |-----------|------------ |--------------------------------------------------|
| `--state-file`      | string    | -           | Path to Terraform .tfstate                       |
| `--hcl-dir`         | string    | -           | Path to Terraform HCL directory                  |
| `--attributes`      | string    | -           | Comma-separated attributes to check              |
| `--output`          | string    | `console`   | Output format (`console`, `json`, `both`)        |
| `--output-file`     | string    | -           | File to save report (if JSON)                    |
| `--parallel-checks` | number    | 0           | No of concurrent checks                          |
| `--log-level`       | string    | `INFO`      | Determines the max log level                     |
| `--source-of-truth` | string    | `terraform` | AWS or Terraform                                 |


### Examples
**Non-Drifted**

JSON output to stdout
```bash
$ go run cmd/drift-detector/main.go detect --state-file=terraform/terraform.tfstate --attributes=instance_type,tags,vpc_security_group_ids --output=json

2025-05-01T10:43:19.113+0100 [WARN]  drift-detector: No configuration file found, will check for .envrc and environment variables
2025-05-01T10:43:19.113+0100 [INFO]  drift-detector: Loading configuration from .envrc file: /Users/bigvic/github.com/victor-devv/ec2-drift-detector/.envrc
2025-05-01T10:43:19.114+0100 [INFO]  drift-detector: Configuration loaded successfully
2025-05-01T10:43:19.114+0100 [INFO]  drift-detector: Creating in-memory drift repository
2025-05-01T10:43:19.114+0100 [INFO]  drift-detector: Reporters created successfully
2025-05-01T10:43:19.114+0100 [INFO]  drift-detector: Using LocalStack endpoint: http://localhost:4566: component=aws-client
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: AWS client initialized successfully: component=aws-client
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: AWS provider initialized
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: Terraform provider initialized
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: Creating drift detector with source of truth: terraform
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: Drift detector created successfully
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: Updating reporters: component=drift-detector
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: Detecting drift for all instances: component=cli-handler
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: Detecting and reporting drift for all instances: component=drift-detector
2025-05-01T10:43:19.150+0100 [INFO]  drift-detector: Detecting drift for all instances: component=drift-detector
2025-05-01T10:43:19.151+0100 [INFO]  drift-detector: Listing instances from Terraform: component=terraform-client
2025-05-01T10:43:19.151+0100 [INFO]  drift-detector: Parsing Terraform state file: terraform/terraform.tfstate: component=terraform-state
2025-05-01T10:43:19.151+0100 [INFO]  drift-detector: Listing all EC2 instances: component=aws-ec2
2025-05-01T10:43:19.151+0100 [INFO]  drift-detector: Successfully parsed Terraform state file with 5 resources: component=terraform-state
2025-05-01T10:43:19.151+0100 [INFO]  drift-detector: Extracting EC2 instances from Terraform state: component=terraform-state
2025-05-01T10:43:19.151+0100 [INFO]  drift-detector: Found 2 EC2 instances in Terraform state: component=terraform-state
2025-05-01T10:43:19.169+0100 [INFO]  drift-detector: Found 2 EC2 instances: component=aws-ec2
2025-05-01T10:43:19.169+0100 [INFO]  drift-detector: Detecting drift for instance i-471adec4374c632bf: component=drift-detector
2025-05-01T10:43:19.169+0100 [INFO]  drift-detector: Detecting drift for instance i-c9b8969cbf6dc304d: component=drift-detector
2025-05-01T10:43:19.169+0100 [INFO]  drift-detector: Reporting drift for 2 instances: component=drift-detector
2025-05-01T10:43:19.169+0100 [INFO]  drift-detector: Reporting drift for 2 instances to JSON file: component=json-reporter
{
  "timestamp": "2025-05-01T10:43:19.169499+01:00",
  "total_instances": 2,
  "drifted_count": 0,
  "results": [
    {
      "id": "2005a649-4ec4-4eac-b8da-37578822967b",
      "resource_id": "i-c9b8969cbf6dc304d",
      "resource_type": "aws_instance",
      "source_type": "terraform",
      "timestamp": "2025-05-01T10:43:19.169273+01:00",
      "has_drift": false
    },
    {
      "id": "fda5edf1-93b5-4dc5-8ab9-08219894f79f",
      "resource_id": "i-471adec4374c632bf",
      "resource_type": "aws_instance",
      "source_type": "terraform",
      "timestamp": "2025-05-01T10:43:19.169262+01:00",
      "has_drift": false
    }
  ]
}
2025-05-01T10:43:19.169+0100 [INFO]  drift-detector: Successfully written report to stdout: component=json-reporter
```

Console output
```bash
$ go run cmd/drift-detector/main.go detect --state-file=terraform/terraform.tfstate --attributes=instance_type,tags,vpc_security_group_ids --output=console --output-file=drift_report.json

2025-05-01T10:21:03.307+0100 [WARN]  drift-detector: No configuration file found, will check for .envrc and environment variables
2025-05-01T10:21:03.307+0100 [INFO]  drift-detector: Loading configuration from .envrc file: /Users/bigvic/github.com/victor-devv/ec2-drift-detector/.envrc
2025-05-01T10:21:03.309+0100 [INFO]  drift-detector: Configuration loaded successfully
2025-05-01T10:21:03.309+0100 [INFO]  drift-detector: Creating in-memory drift repository
2025-05-01T10:21:03.309+0100 [INFO]  drift-detector: Reporters created successfully
2025-05-01T10:21:03.309+0100 [INFO]  drift-detector: Using LocalStack endpoint: http://localhost:4566: component=aws-client
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: AWS client initialized successfully: component=aws-client
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: AWS provider initialized
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Terraform provider initialized
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Creating drift detector with source of truth: terraform
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Drift detector created successfully
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Updating reporters: component=drift-detector
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Detecting drift for all instances: component=cli-handler
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Detecting and reporting drift for all instances: component=drift-detector
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Detecting drift for all instances: component=drift-detector
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Listing instances from Terraform: component=terraform-client
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Listing all EC2 instances: component=aws-ec2
2025-05-01T10:21:03.349+0100 [INFO]  drift-detector: Parsing Terraform state file: terraform/terraform.tfstate: component=terraform-state
2025-05-01T10:21:03.351+0100 [INFO]  drift-detector: Successfully parsed Terraform state file with 5 resources: component=terraform-state
2025-05-01T10:21:03.351+0100 [INFO]  drift-detector: Extracting EC2 instances from Terraform state: component=terraform-state
2025-05-01T10:21:03.351+0100 [INFO]  drift-detector: Found 2 EC2 instances in Terraform state: component=terraform-state
2025-05-01T10:21:03.361+0100 [INFO]  drift-detector: Found 2 EC2 instances: component=aws-ec2
2025-05-01T10:21:03.361+0100 [INFO]  drift-detector: Detecting drift for instance i-c9b8969cbf6dc304d: component=drift-detector
2025-05-01T10:21:03.361+0100 [INFO]  drift-detector: Detecting drift for instance i-471adec4374c632bf: component=drift-detector
2025-05-01T10:21:03.361+0100 [INFO]  drift-detector: Reporting drift for 2 instances: component=drift-detector
&{0x1400021d0e0 true}
2025-05-01T10:21:03.361+0100 [INFO]  drift-detector: Reporting drift for 2 instances: component=console-reporter
=== Drift Detection Summary ===

Number of Instances: 2
Instances with Drift: No (0/2)

No drift detected in any instance.
```

**Drifted**

Console Output
```bash
$ go run cmd/drift-detector/main.go detect --state-file=terraform/terraform.tfstate --attributes=instance_type,tags,vpc_security_group_ids --output=console --output-file=drift_report.json

2025-05-01T10:24:28.805+0100 [WARN]  drift-detector: No configuration file found, will check for .envrc and environment variables
2025-05-01T10:24:28.805+0100 [INFO]  drift-detector: Loading configuration from .envrc file: /Users/bigvic/github.com/victor-devv/ec2-drift-detector/.envrc
2025-05-01T10:24:28.806+0100 [INFO]  drift-detector: Configuration loaded successfully
2025-05-01T10:24:28.806+0100 [INFO]  drift-detector: Creating in-memory drift repository
2025-05-01T10:24:28.806+0100 [INFO]  drift-detector: Reporters created successfully
2025-05-01T10:24:28.806+0100 [INFO]  drift-detector: Using LocalStack endpoint: http://localhost:4566: component=aws-client
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: AWS client initialized successfully: component=aws-client
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: AWS provider initialized
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Terraform provider initialized
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Creating drift detector with source of truth: terraform
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Drift detector created successfully
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Updating reporters: component=drift-detector
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Detecting drift for all instances: component=cli-handler
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Detecting and reporting drift for all instances: component=drift-detector
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Detecting drift for all instances: component=drift-detector
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Listing instances from Terraform: component=terraform-client
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Listing all EC2 instances: component=aws-ec2
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Parsing Terraform state file: terraform/terraform.tfstate: component=terraform-state
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Successfully parsed Terraform state file with 5 resources: component=terraform-state
2025-05-01T10:24:28.852+0100 [INFO]  drift-detector: Extracting EC2 instances from Terraform state: component=terraform-state
2025-05-01T10:24:28.853+0100 [INFO]  drift-detector: Found 2 EC2 instances in Terraform state: component=terraform-state
2025-05-01T10:24:28.869+0100 [INFO]  drift-detector: Found 2 EC2 instances: component=aws-ec2
2025-05-01T10:24:28.869+0100 [INFO]  drift-detector: Detecting drift for instance i-471adec4374c632bf: component=drift-detector
2025-05-01T10:24:28.869+0100 [INFO]  drift-detector: Detecting drift for instance i-c9b8969cbf6dc304d: component=drift-detector
2025-05-01T10:24:28.869+0100 [INFO]  drift-detector: Detected 1 drifted attributes for instance i-c9b8969cbf6dc304d: component=drift-detector
2025-05-01T10:24:28.869+0100 [INFO]  drift-detector: Reporting drift for 2 instances: component=drift-detector
2025-05-01T10:24:28.869+0100 [INFO]  drift-detector: Reporting drift for 2 instances: component=console-reporter
=== Drift Detection Summary ===

Number of Instances: 2
Instances with Drift: Yes (1/2)

=== Instances with Drift ===

Instance ID          Drifted Attributes  Timestamp
-----------          ------------------  ---------
i-c9b8969cbf6dc304d  instance_type       2025-05-01T10:24:28+01:00

Use 'drift-detector show <instance-id>' to see detailed drift information for a specific instance.   
```

JSON output to stdout
```bash
$ go run cmd/drift-detector/main.go detect --state-file=terraform/terraform.tfstate --attributes=instance_type,tags,vpc_security_group_ids --output=json

2025-05-01T10:41:50.582+0100 [WARN]  drift-detector: No configuration file found, will check for .envrc and environment variables
2025-05-01T10:41:50.582+0100 [INFO]  drift-detector: Loading configuration from .envrc file: /Users/bigvic/github.com/victor-devv/ec2-drift-detector/.envrc
2025-05-01T10:41:50.583+0100 [INFO]  drift-detector: Configuration loaded successfully
2025-05-01T10:41:50.583+0100 [INFO]  drift-detector: Creating in-memory drift repository
2025-05-01T10:41:50.583+0100 [INFO]  drift-detector: Reporters created successfully
2025-05-01T10:41:50.583+0100 [INFO]  drift-detector: Using LocalStack endpoint: http://localhost:4566: component=aws-client
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: AWS client initialized successfully: component=aws-client
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: AWS provider initialized
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: Terraform provider initialized
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: Creating drift detector with source of truth: terraform
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: Drift detector created successfully
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: Updating reporters: component=drift-detector
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: Detecting drift for all instances: component=cli-handler
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: Detecting and reporting drift for all instances: component=drift-detector
2025-05-01T10:41:50.602+0100 [INFO]  drift-detector: Detecting drift for all instances: component=drift-detector
2025-05-01T10:41:50.603+0100 [INFO]  drift-detector: Listing instances from Terraform: component=terraform-client
2025-05-01T10:41:50.603+0100 [INFO]  drift-detector: Parsing Terraform state file: terraform/terraform.tfstate: component=terraform-state
2025-05-01T10:41:50.603+0100 [INFO]  drift-detector: Listing all EC2 instances: component=aws-ec2
2025-05-01T10:41:50.603+0100 [INFO]  drift-detector: Successfully parsed Terraform state file with 5 resources: component=terraform-state
2025-05-01T10:41:50.603+0100 [INFO]  drift-detector: Extracting EC2 instances from Terraform state: component=terraform-state
2025-05-01T10:41:50.603+0100 [INFO]  drift-detector: Found 2 EC2 instances in Terraform state: component=terraform-state
2025-05-01T10:41:50.613+0100 [INFO]  drift-detector: Found 2 EC2 instances: component=aws-ec2
2025-05-01T10:41:50.613+0100 [INFO]  drift-detector: Detecting drift for instance i-471adec4374c632bf: component=drift-detector
2025-05-01T10:41:50.613+0100 [INFO]  drift-detector: Detecting drift for instance i-c9b8969cbf6dc304d: component=drift-detector
2025-05-01T10:41:50.613+0100 [INFO]  drift-detector: Detected 1 drifted attributes for instance i-c9b8969cbf6dc304d: component=drift-detector
2025-05-01T10:41:50.613+0100 [INFO]  drift-detector: Reporting drift for 2 instances: component=drift-detector
2025-05-01T10:41:50.614+0100 [INFO]  drift-detector: Reporting drift for 2 instances to JSON file: component=json-reporter
{
  "timestamp": "2025-05-01T10:41:50.614026+01:00",
  "total_instances": 2,
  "drifted_count": 1,
  "results": [
    {
      "id": "a099518c-07b0-4bb2-9b75-cca232157d01",
      "resource_id": "i-c9b8969cbf6dc304d",
      "resource_type": "aws_instance",
      "source_type": "terraform",
      "timestamp": "2025-05-01T10:41:50.613743+01:00",
      "has_drift": true,
      "drifted_attributes": {
        "instance_type": {
          "path": "instance_type",
          "source_value": "t2.micro",
          "target_value": "t3.micro",
          "changed": true
        }
      }
    },
    {
      "id": "848a1f72-53e7-449b-a727-c6e2d0094f50",
      "resource_id": "i-471adec4374c632bf",
      "resource_type": "aws_instance",
      "source_type": "terraform",
      "timestamp": "2025-05-01T10:41:50.61373+01:00",
      "has_drift": false
    }
  ]
}
2025-05-01T10:41:50.614+0100 [INFO]  drift-detector: Successfully written report to stdout: component=json-reporter 
```

---

### 🧾 Sample AWS EC2 Response (JSON)

This is an example of how an EC2 instance might look in the internal model after being fetched from AWS and serialized as JSON:

```json
{
  "id": "i-0abc1234567890def",
  "instance_type": "t3.micro",
  "ami": "ami-0abcdef1234567890",
  "subnet_id": "subnet-0123456789abcdef0",
  "vpc_id": "vpc-0a1b2c3d4e5f6g7h8",
  "security_group_ids": ["sg-0123456789abcdef0"],
  "security_group_names": ["localstack-sg"],
  "key_name": "my-keypair",
  "iam_role": "ec2-readonly",
  "public_ip": "54.123.45.67",
  "private_ip": "10.0.1.23",
  "tags": {
    "Name": "LocalStack EC2",
    "Env": "dev"
  },
  "root_volume_size": 8,
  "root_volume_type": "gp2",
  "user_data": "IyEvYmluL2Jhc2gKZWNobyBIZWxsbyBXb3JsZA==",
  "ebs_optimized": false,
  "source_dest_check": true,
  "monitoring_enabled": false,
  "termination_protection": false
}
```

---

## 🧱 Design & Architecture

The application follows Domain-Driven Design principles:

- **Domain Layer**: Core business logic and models
- **Application Layer**: Orchestration of domain services
- **Infrastructure Layer**: External dependencies (AWS, Terraform)
- **Presentation Layer**: CLI and reporting interfaces

### Project Structure

```
terraform-drift-detector/
├── cmd/                    # Command-line entry points
├── internal/               # Internal packages
│   ├── app/                # Application services
│   ├── common/             # Common utilities
│   ├── config/             # Configuration
│   ├── container/          # DI Container
│   ├── domain/             # Domain models and services
│   ├── factory/            # App component initialization
│   ├── infrastructure/     # External dependencies
│   └── presentation/       # User interfaces
├── pkg/                    # Public packages
├── terraform/              # Terraform manifests for mock AWS resources
├── config.yaml.example     # Sample yaml configuration
├── .envrc.example          # Sample .envrc configuration
├── docker-compose.yml      # Docker Compose configuration
├── Dockerfile              # Docker build file
├── go.mod                  # Go modules
├── Makefile                # Build and development commands
└── README.md               # This file
```

## 🧱 Architecture Diagram (Logical)

```
                 +--------------------+
                 |      CLI (cmd)     |
                 +--------------------+
                           |
      +--------------------+---------------------+
      |                    |                     |
      v                    v                     v
+---------------+  +----------------+   +--------------------+
| Terraform     |  | AWS EC2 Client |   |  Drift Detector    |
| Parser        |  | (internal/aws) |   |  (internal/detector)|
| (terraform/)  |  +----------------+   +--------------------+
+---------------+           |                    |
                            +--------------------+
                                      |
                             +-------------------+
                             |     Reporter      |
                             | (internal/reporter)|
                             +-------------------+
```

## 🔍 Description

- **CLI (cmd/)**: Entry point that parses arguments and triggers drift detection.
- **Terraform Parser**: Parses `.tfstate` or `.tf` HCL files into internal models.
- **AWS EC2 Client**: Fetches real-time EC2 configurations from AWS.
- **Drift Detector**: Compares Terraform vs AWS instance data.
- **Reporter**: Outputs results to JSON or console.
---

### Approach
 - Domain driven design with the core domain as "drift detection"
 - Comprehensive error system and management
 - Flexible configuration management
 - SOLID implementation
 - Consistent DriftResult model for easy formatting
 - JSON-encoded reports for downstream processing

### Trade-Offs
 - Limited to EC2 drift only for now (no ELBs, RDS, etc.)
 - Implemented an in-memory repository for drift results (persistence over performance)

### ⚠️ Challenges Faced
 - Issues parsing HCL configurations which use variables
 - Terraform state's nested and sometimes inconsistent structure
 - Handling differences in how AWS and Terraform express tags
 - Simulating real AWS EC2 behavior in LocalStack
 - Balancing concurrency with predictable logging and output

### 🚀 Future Improvements
 - Implement persistent storage for drift results to enable historical tracking and trend analysis
 - Extend drift detection to other AWS resources (e.g., S3, RDS)
 - Implement integrations with notification systems for automated alerts.
 - Add capability to suggest Terraform commands to resolve detected drift.
 - Support distributed operation for very large infrastructures, perhaps using a work queue.
 - Add encryption, authentication, and authorization for secure enterprise deployment.

---

### Author

Victor Ikuomola
GitHub: @victor-devv
