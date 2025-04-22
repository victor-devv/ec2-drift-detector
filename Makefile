MAIN := cmd/drift-detector/main.go
ATTRIBUTES := instance_type,tags,vpc_security_group_ids
OUTPUT_FORMAT := json
OUTPUT_FILE := drift_report.json

.PHONY: run
run:
	go run $(MAIN) \
		--state-file=$(TERRAFORM_STATE_FILE) \
		--attributes=$(ATTRIBUTES) \
		--output=$(OUTPUT_FORMAT) \
		--output-file=$(OUTPUT_FILE)

.PHONY: build
build:
	go build -o $(APP_NAME) $(MAIN)

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	set -a && source .envrc && set +a && docker run --rm \
		-v $(PWD):/app \
		-e CONCURRENT=$(CONCURRENT) \
		-e LOG_LEVEL=$(LOG_LEVEL) \
		-e AWS_DEFAULT_REGION=$(AWS_DEFAULT_REGION) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		-e AWS_EC2_ENDPOINT=$(AWS_EC2_ENDPOINT) \
		-e TERRAFORM_STATE_FILE=$(TERRAFORM_STATE_FILE) \
		$(APP) --state-file=$(TERRAFORM_STATE_FILE) --attributes=$(ATTRIBUTES)

.PHONY: localstack-up
localstack-up:
	docker-compose up -d

.PHONY: localstack-down
localstack-down:
	docker-compose down

.PHONY: tf-init
tf-init:
	cd terraform && terraform init

.PHONY: tf-plan
tf-plan:
	cd terraform && terraform plan -var-file="config.tfvars" -out="tfplan"

.PHONY: tf-apply
tf-apply:
	cd terraform && terraform apply tfplan -auto-approve

.PHONY: tf-destroy
tf-destroy:
	cd terraform && terraform destroy -auto-approve

.PHONY: test
test:
	go test ./... -v

.PHONY: cover
cover:
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

.PHONY: cover-html
cover-html:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
	@echo "✔️  View coverage report at: coverage.html"

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
