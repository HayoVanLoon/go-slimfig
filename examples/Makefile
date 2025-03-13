GO := $(shell which go)

# A demonstration of basic file-based resolvers. The YAML, which is loaded
# after the JSON, modifies some of the values from there.
example-file-based:
	$(GO) run filebased/*.go \
		./config.json \
		./config.yaml

# Like the file-based example, but overwriting the 'key' again with a value
# taken from the environment.
example-env:
	EX_key="key from env" $(GO) run filebased/*.go \
		./config.json \
		./config.yaml

# Uses the Google Cloud Secret Manager resolver for one of the references.
example-gcp:
ifndef PROJECT_ID
	$(error missing GCP_PROJECT_ID)
endif
ifndef SECRET_ID
	$(error missing SECRET_ID)
endif
	$(GO) run gcp/*.go \
		./config.json \
		'gcp-secretmanager://projects/$(PROJECT_ID)/secrets/$(SECRET_ID)'

# Uses the AWS Secrets Manager resolver for one of the references.
example-aws:
ifndef SECRET_ID
	$(error missing SECRET_ID)
endif
	$(GO) run aws/*.go \
		./config.json \
		'aws-secretsmanager://$(SECRET_ID)'
