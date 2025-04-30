.PHONY: build test test-report clean run run-binary run-binary-server run-binary-config-show run-binary-config-reload run-config-show run-config-reload run-server localstack terraform docker-build docker-run tf-init tf-plan tf-apply

GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=drift-detector
BINARY_UNIX=$(BINARY_NAME)_unix
LDFLAGS=-ldflags "-X main.Version=0.1.0"

MAIN=cmd/drift-detector/main.go
ATTRIBUTES=instance_type,tags,vpc_security_group_ids
STATE_FILE=terraform/terraform.tfstate
OUTPUT_FORMAT=json
OUTPUT_FILE=drift_report.json

DOCKER_IMAGE=ec2-drift-detector

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/drift-detector

test: 
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

test-report:
	$(GOTEST) ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
	@echo "✔️  View coverage report at: coverage.html"

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out
	rm -f coverage.html

run:
	$(GORUN) $(MAIN) detect --state-file=$(STATE_FILE) --attributes=$(ATTRIBUTES) --output=$(OUTPUT_FORMAT) --output-file=$(OUTPUT_FILE)

run-config-show:
	$(GORUN) $(MAIN) config show

run-config-reload:
	$(GORUN) $(MAIN) config reload

run-server:
	$(GORUN) $(MAIN) server

run-binary:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/drift-detector
	./$(BINARY_NAME) detect --state-file=$(STATE_FILE) --attributes=$(ATTRIBUTES) --output=$(OUTPUT_FORMAT) --output-file=$(OUTPUT_FILE)

run-binary-server:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/drift-detector
	./$(BINARY_NAME) server

run-binary-config-show:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/drift-detector
	./$(BINARY_NAME) config show

run-binary-config-reload:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/drift-detector
	./$(BINARY_NAME) config reload

# Manage dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

# Run localstack for development
localstack-up:
	docker-compose up -d localstack

localstack-down:
	docker-compose down

docker-build:
	docker build -t $(DOCKER_IMAGE) .
	
docker-run:
	docker run --rm -it $(DOCKER_IMAGE)

# Start the complete stack with Docker Compose
start:
	docker-compose up -d

stop:
	docker-compose down

tf-init:
	cd terraform && terraform init

tf-plan:
	cd terraform && terraform plan -var-file="config.tfvars" -out="tfplan"

tf-apply:
	cd terraform && terraform apply -auto-approve tfplan

tf-destroy:
	cd terraform && terraform destroy -auto-approve

.PHONY: godoc-install
godoc-install:
	go install golang.org/x/tools/cmd/godoc@latest

.PHONY: godoc
godoc:
	godoc -http=:6060

.PHONY: pkgsite-install
pkgsite-install:
	go install golang.org/x/pkgsite/cmd/pkgsite@latest

.PHONY: pkgsite
pkgsite:
	pkgsite -http=localhost:7070
