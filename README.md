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

🧪 Run Drift Detection (edit cli params in make file or .envrc)

```bash
make run
```

Or Manually

```bash
./ec2-drift-detector \
  --state-file=internal/terraform/terraform.tfstate \
  --attributes=instance_type,tags,ami \
  --output=json \
  --output-file=drift-report.json
```

### 🧪 Run Tests

```bash
make test
make cover-summary
make cover-html
```

---

### 🧭 CLI Usage Example

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
INFO[0000] Drift detection completed
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

### Project Structure

| Directory          | Description                          |
|--------------------|--------------------------------------|
| `/cmd/`            | CLI entrypoint                       |
| `/internal/aws/`   | AWS SDK wrappers                     |
| `/internal/config/`| Environment and validation config    |
| `/internal/detector/` | Core drift detection logic        |
| `/internal/terraform/` | Terraform state parser           |
| `/internal/cli/`   | CLI argument parser                  |
| `/internal/reporter/` | Console/JSON report output       |
| `/pkg/logger/`     | Logrus-based logger                  |
| `/pkg/utils/`      | Miscellaneous utilities (file helpers, etc.) |

## 🧱 Architecture Diagram (Logical)
```mermaid
flowchart TD
    A[CLI (cmd/)] -->|Triggers| B[Terraform Parser\n(internal/terraform)]
    A -->|Triggers| C[AWS EC2 Client\n(internal/aws)]
    A -->|Triggers| D[Drift Detector\n(internal/detector)]
    
    B -->|Parses tfstate| D
    C -->|Fetches live EC2 config| D
    
    D -->|Generates DriftResult[]| E[Reporter\n(internal/reporter)]
    
    subgraph Legend
        F[CLI] --> G[Core Components]
        G --> H[Output]
    end
```

### Approach
 - Interface-driven design for testability
 - Parallelized drift checks using errgroup
 - Consistent DriftResult model for easy formatting
 - JSON-encoded reports for downstream processing

### Trade-Offs
 - Uses Go stdlib flag instead of cobra for simplicity
 - Limited to EC2 drift only for now (no ELBs, RDS, etc.)
 - Fails to parse HCL configs with variables

### ⚠️ Challenges Faced
 - Issues parsing HCL configurations which use variables
 - Terraform state's nested and sometimes inconsistent structure
 - Handling differences in how AWS and Terraform express tags
 - Simulating real AWS EC2 behavior in LocalStack
 - Balancing concurrency with predictable logging and output

### 🚀 Future Improvements
 - Add support for HCL .tf parsing with variables
 - Extend drift detection to other AWS resources (e.g., S3, RDS)
 - Use cobra or urfave/cli for multi-command CLI (scan, report, etc.)
 - Upload reports to S3 or Slack webhook
 - Web-based dashboard for viewing drift results over time
 - GitHub Actions integration for CI-based drift detection

---

### Author

Victor Ikuomola
GitHub: @victor-devv
