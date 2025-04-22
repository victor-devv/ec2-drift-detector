build:
	docker build -t $(APP) .

run:
	set -a && source .envrc && set +a && docker run --rm \
		-v $(PWD):/app \
		-e CONCURRENT=$(CONCURRENT) \
		-e LOG_LEVEL=$(LOG_LEVEL) \
		-e AWS_DEFAULT_REGION=$(AWS_DEFAULT_REGION) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		-e AWS_EC2_ENDPOINT=$(AWS_EC2_ENDPOINT) \
		-e TERRAFORM_STATE_FILE=$(TERRAFORM_STATE_FILE) \
		$(APP) --state-file=$(TERRAFORM_STATE_FILE) --attributes=instance_type,tags,security_groups

localstack-up:
	docker-compose up -d

localstack-down:
	docker-compose down

tf-init:
	cd terraform && terraform init

tf-apply:
	cd terraform && terraform apply -auto-approve

tf-destroy:
	cd terraform && terraform destroy -auto-approve
