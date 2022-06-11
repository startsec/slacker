SHELL = /bin/bash
SRC = /tmp/build
REV = $(shell git rev-parse HEAD | cut -c1-6 )

LAMBDA_DEPLOY_BUCKET = my-s3-bucket
LAMBDA_S3_KEY = slacker/slacker_${REV}.zip
LAMBDA_FUNCTION_NAME = slacker

build:
	mkdir -p $(SRC)
	GOOS=linux go build main.go
	zip $(SRC)/$(LAMBDA_FUNCTION_NAME)_$(REV).zip main

deploy:
	aws s3 cp \
		--metadata SHA256Sum=$(shell sha256sum $(SRC)/$(LAMBDA_FUNCTION_NAME)_$(REV).zip | cut -d ' ' -f1) \
		"$(SRC)/$(LAMBDA_FUNCTION_NAME)_$(REV).zip" \
		"s3://${LAMBDA_DEPLOY_BUCKET}/${LAMBDA_S3_KEY}"

publish: deploy
	aws lambda --region us-east-1 update-function-code \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--s3-bucket $(LAMBDA_DEPLOY_BUCKET) \
		--s3-key $(LAMBDA_S3_KEY) \
		--publish