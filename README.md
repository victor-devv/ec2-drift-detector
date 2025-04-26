# EC2 Drift Detector

A Go-based CLI tool that detects configuration drift between **AWS EC2 instances** and their corresponding **Terraform definitions**.

---

## 🧠 Project Overview

The EC2 Drift Detector identifies mismatches between the actual state of EC2 instances in AWS and their declared configurations in Terraform. This tool is helpful for:

- Infrastructure audits
- Change detection
- CI/CD integrity checks
- Compliance verification

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
- [Terraform](https://developer.hashicorp.com/terraform/downloads)
- [Direnv](https://direnv.net/)

### 🔧 Build from source

```bash
git clone https://github.com/victor-devv/ec2-drift-detector.git
cd ec2-drift-detector
cp .envrc.example .envrc
direnv allow
make build
```

🐳 Start LocalStack (mock AWS)

```bash
make localstack-up
make tf-init
make tf-apply
```

🧪 Run Drift Detection (built binary)

```bash
make run-binary
```

Or using go run

```bash
make run
```

### 🧪 Run Tests

```bash
make test
make cover-summary
make cover-html
```

### Running the Application

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

## Configuration

The application can be configured through:

1. Configuration files (YAML)
2. Environment variables via .env or .envrc
3. Command-line flags

### Samply config.yaml

```yaml
app:
  log_level: INFO
  json_logs: false
  schedule_expression: "0 */6 * * *"

aws:
  region: us-west-2
  profile: default
  use_localstack: false

terraform:
  state_file: terraform.tfstate
  # Alternatively, use HCL files:
  # hcl_dir: terraform/
  # use_hcl: true

drift_detection:
  source_of_truth: terraform
  attributes:
    - instance_type
    - ami
    - vpc_security_group_ids
    - tags
  parallel_checks: 5
  timeout_seconds: 60

reporter:
  type: both  # console, json, or both
  output_file: drift-report.json
  pretty_print: true
```
### Environment Variables

Environment variables are prefixed with `DRIFT_` and use underscores:

```bash
export DRIFT_APP_LOG_LEVEL=DEBUG
export DRIFT_AWS_REGION=us-west-2
export DRIFT_AWS_USE_LOCALSTACK=true
export DRIFT_TERRAFORM_STATE_FILE=terraform.tfstate
export DRIFT_DRIFT_DETECTION_SOURCE_OF_TRUTH=terraform
```

### Command-Line Flags

```bash
./drift-detector detect --state-file=terraform.tfstate

---

### 🧭 CLI Usage Examples

**Non-Drifted**

```bash
$ go run cmd/drift-detector/main.go \
  --state-file=terraform.tfstate \
  --attributes=instance_type,tags,ami \
  --output=json \
  --output-file=drift-report.json \
  --verbose

INFO[0000] Starting drift detection
INFO[0000] Terraform path: terraform.tfstate
INFO[0000] Attributes to check: [instance_type tags ami]
INFO[0000] Output format: json
INFO[0000] Running concurrent drift detection
INFO[0000] Concurrent drift detection complete: 1 instance(s) compared 
{
  "summary": {
    "totalResources": 1,
    "driftedResources": 0,
    "nonDriftedResources": 1
  },
  "results": [
    {
      "resourceId": "i-3710997b49f48cdc3",
      "resourceType": "aws_instance",
      "inTerraform": true,
      "inAWS": true,
      "drifted": false,
      "attribute_diffs": []
    }
  ]
}
INFO[0000] Drift detection completed
```

```bash
$ go run cmd/drift-detector/main.go \
  --state-file=terraform.tfstate \
  --attributes=instance_type,tags,ami \
  --output=console \
  --verbose

INFO[0000] Starting drift detection                     
INFO[0000] Terraform path: terraform/terraform.tfstate  
INFO[0000] Attributes to check: [instance_type tags vpc_security_group_ids] 
INFO[0000] Output format: console                       
INFO[0000] Running concurrent drift detection           
INFO[0000] Concurrent drift detection complete: 1 instance(s) compared 
✓ No drift detected across 1 resource(s)

RESOURCE ID          TYPE          STATUS   DETAILS
----------           ----          ------   -------
i-3710997b49f48cdc3  aws_instance  OK  -
INFO[0000] Drift detection completed  
```

**Drifted**

```bash
$ go run cmd/drift-detector/main.go \
  --state-file=terraform.tfstate \
  --attributes=instance_type,tags,ami \
  --output=console \
  --verbose
  
INFO[0000] Starting drift detection                     
INFO[0000] Terraform path: terraform/terraform.tfstate  
INFO[0000] Attributes to check: [instance_type tags vpc_security_group_ids] 
INFO[0000] Output format: console                       
INFO[0000] Running concurrent drift detection           
INFO[0000] Concurrent drift detection complete: 1 instance(s) compared 
✗ Drift detected in 1 out of 1 resource(s)

RESOURCE ID          TYPE          STATUS        DETAILS
----------           ----          ------        -------
i-3710997b49f48cdc3  aws_instance  DRIFTED  instance_type: AWS='t3.micro', TF='t2.micro'
INFO[0000] Drift detection completed    
```

```bash
$ go run cmd/drift-detector/main.go \
  --state-file=terraform.tfstate \
  --attributes=instance_type,tags,ami \
  --output=console \
  --verbose

INFO[0000] Starting drift detection                     
INFO[0000] Terraform path: terraform/terraform.tfstate  
INFO[0000] Attributes to check: [instance_type tags vpc_security_group_ids] 
INFO[0000] Output format: json                          
INFO[0000] Running concurrent drift detection           
INFO[0000] Concurrent drift detection complete: 1 instance(s) compared 
{
  "summary": {
    "totalResources": 1,
    "driftedResources": 1,
    "nonDriftedResources": 0
  },
  "results": [
    {
      "resourceId": "i-3710997b49f48cdc3",
      "resourceType": "aws_instance",
      "inTerraform": true,
      "inAWS": true,
      "drifted": true,
      "attribute_diffs": [
        {
          "attribute_name": "instance_type",
          "aws_value": "t3.micro",
          "terraform_value": "t2.micro",
          "is_complex": false
        }
      ]
    }
  ]
}
INFO[0000] Drift detection completed  
```

---

## 🧾 Sample AWS EC2 Response (JSON)

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

### ⚙️ Configuration

**Set via CLI flags or environment variables**:

| Flag           | Type      | Default    | Description                                      |
|----------------|-----------|------------|--------------------------------------------------|
| `--state-file` | string    | -          | Path to Terraform .tfstate                       |
| `--attributes` | string    | -          | Comma-separated attributes to check              |
| `--output`     | string    | `console`  | Output format (`console`, `json`)                |
| `--output-file`| string    | -          | File to save report (if JSON)                    |
| `--concurrent` | bool      | `false`    | Run checks concurrently                          |
| `--verbose`    | bool      | `false`    | Enable debug logs                                |

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
│   ├── domain/             # Domain models and services
│   ├── infrastructure/     # External dependencies
│   └── presentation/       # User interfaces
├── pkg/                    # Public packages
├── test/                   # Tests and fixtures
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
